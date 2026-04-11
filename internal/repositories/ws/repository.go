package ws

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/libs/logging"
	"github.com/gorilla/websocket"
)

const (
	readMessagesBufferSize = 100
	readErrorsBufferSize   = 1

	writeDeadline = 2 * time.Second

	accessTokenCookieName = "accessToken"
)

type Repository struct {
	baseURL      string
	logger       logging.Logger
	ws           *websocket.Conn
	mu           sync.Mutex
	messagesChan chan *domains.Message // Буферизированный канал для входящих сообщений
	errChan      chan error            // Канал для критических ошибок чтения
}

func New(baseURL string, logger logging.Logger) *Repository {
	return &Repository{
		baseURL: baseURL,
		logger:  logger,
	}
}

func (r *Repository) Connect(ctx context.Context, accessToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Безопасный вызов уже инициализированного вебсокета
	if r.ws != nil {
		return nil
	}

	// Создаем каналы
	r.messagesChan = make(
		chan *domains.Message,
		readMessagesBufferSize,
	) // Буфер, чтобы не блокировать чтение
	r.errChan = make(chan error, readErrorsBufferSize)

	header := http.Header{}
	header.Add(
		common.CookieHeaderName,
		fmt.Sprintf("%s=%s", accessTokenCookieName, accessToken),
	)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, r.baseURL+"/ws", header)
	if err != nil {
		return err
	}

	r.ws = conn

	// Запускаем чтение из сокета
	go r.readLoop()

	return nil
}

func (r *Repository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Безопасное закрытие вебсокета
	if r.ws == nil {
		return nil
	}

	err := r.ws.Close()

	r.ws = nil

	return err
}

func (r *Repository) ReadMessage(ctx context.Context) (*domains.Message, error) {
	if r.ws == nil {
		return nil, fmt.Errorf("%w: connection was not enstablished", customerrors.ErrWebsocket)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-r.errChan:
		return nil, err
	case msg, ok := <-r.messagesChan:
		// Канал закрыт
		if !ok {
			// Проверяем, есть ли ошибка, если нет - отдаем дефолтную
			select {
			case err := <-r.errChan:
				return nil, err
			default:
				return nil, customerrors.ErrWebsocketClosed
			}
		}

		return msg, nil
	}
}

func (r *Repository) WriteMessage(ctx context.Context, message domains.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, не закрыт ли сокет (под мьютексом)
	if r.ws == nil {
		return fmt.Errorf("%w: connection not established", customerrors.ErrWebsocket)
	}

	// Устанавливаем дедлайн из контекста.
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(writeDeadline))
	defer cancelFunc()

	if deadline, ok := ctx.Deadline(); ok {
		if err := r.ws.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf(
				"%w: failed to set websocket write deadline",
				customerrors.ErrWebsocket,
			)
		}
	}

	// Сбрасываем дедлайн по выходу, чтобы сокет был "чистым" для других операций
	defer func() {
		if err := r.ws.SetWriteDeadline(time.Time{}); err != nil {
			logging.LogError(r.logger, "failed to set ws write deadline", err)
		}
	}()

	// Пишем в сокет (он прервется сам, если наступит deadline)
	if err := r.ws.WriteJSON(message); err != nil {
		if websocket.IsCloseError(
			err,
			websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure,
		) ||
			errors.Is(err, net.ErrClosed) {
			return customerrors.ErrWebsocketClosed
		}

		return fmt.Errorf("%w: %w", customerrors.ErrWebsocket, err)
	}

	return nil
}

func (r *Repository) readLoop() {
	defer close(r.messagesChan)
	defer close(r.errChan)

	for {
		var msg domains.Message
		if err := r.ws.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				r.errChan <- customerrors.ErrWebsocketClosed
			} else {
				r.errChan <- fmt.Errorf("%w: %w", customerrors.ErrWebsocket, err)
			}

			if err = r.Close(); err != nil {
				logging.LogError(r.logger, "failed to close ws connection", err)
			}

			return
		}

		r.messagesChan <- &msg
	}
}
