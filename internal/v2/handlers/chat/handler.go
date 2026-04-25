package chat

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	refreshTokensInterval = 1 * time.Minute
	updateChatsInterval   = 5 * time.Second

	chatsUpdatedEventName = "chats_updated"
	newMessageEventName   = "new_message"
)

type Handler struct {
	useCases         interfaces.UseCases
	errorsMapper     interfaces.ErrorsMapper
	logger           logging.Logger
	validationConfig config.ValidationConfig

	wailsCtx                context.Context
	goroutinesCtx           context.Context
	goroutinesCtxCancelFunc context.CancelFunc
	wg                      sync.WaitGroup
}

func New(
	useCases interfaces.UseCases,
	errorsMapper interfaces.ErrorsMapper,
	logger logging.Logger,
	validationConfig config.ValidationConfig,
) *Handler {
	return &Handler{
		useCases:         useCases,
		errorsMapper:     errorsMapper,
		logger:           logger,
		validationConfig: validationConfig,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) GetCurrentUser() (*domains.User, error) {
	ctx := context.Background()

	return h.useCases.GetCurrentUser(ctx)
}

func (h *Handler) GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error) {
	ctx := context.Background()

	return h.useCases.GetUserChats(ctx, pagination)
}

func (h *Handler) GetChatMessages(
	chatID uint64,
	pagination *domains.Pagination,
) ([]domains.Message, error) {
	ctx := context.Background()

	return h.useCases.GetChatMessages(ctx, chatID, pagination)
}

func (h *Handler) SendMessage(chatID uint64, text string) error {
	ctx := context.Background()

	sender, err := h.useCases.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	message := domains.Message{
		ChatID: chatID,
		Text:   text,
		Sender: domains.User{
			ID: sender.ID,
		},
		CreatedAt: time.Now().In(common.Timezone),
		UpdatedAt: time.Now().In(common.Timezone),
		IsRead:    true, // Сообщение прочитано для отправителя
	}

	return h.useCases.SendMessage(ctx, message)
}

func (h *Handler) StartListening() {
	h.goroutinesCtx, h.goroutinesCtxCancelFunc = context.WithCancel(context.Background())

	// TODO Wails CLI v2.12.0 не умеет в wg.Go(), поэтому надо будет обновить в дальнейшем

	// Запускаем горутину для чтения сообщений
	h.wg.Add(1)

	go h.readMessages()

	// Запускаем обновление токенов
	h.wg.Add(1)

	go h.refreshTokens()

	// Запускаем обновление чатов
	h.wg.Add(1)

	go h.updateChats()
}

func (h *Handler) StopListening() {
	if h.goroutinesCtxCancelFunc != nil {
		h.goroutinesCtxCancelFunc()
	}

	h.wg.Wait()
}

func (h *Handler) readMessages() {
	defer h.wg.Done()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		default:
			message, err := h.useCases.ReadMessage(h.goroutinesCtx)
			if err != nil {
				// Соккет закрыт, отключаем горутину
				if errors.Is(err, customerrors.ErrWebsocketClosed) {
					logging.LogInfo(
						h.logger,
						"startReadMessagesGoroutine завершена из-за закрытия соединения",
					)

					return
				}

				logging.LogErrorContext(
					h.goroutinesCtx,
					h.logger,
					"Не удалось прочитать сообщение",
					err,
				)

				continue
			}

			// Отправляем событие на фронтенд
			runtime.EventsEmit(h.wailsCtx, newMessageEventName, message)
		}
	}
}

func (h *Handler) refreshTokens() {
	defer h.wg.Done()

	ticker := time.NewTicker(refreshTokensInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		case <-ticker.C:
			_, _ = h.useCases.RefreshTokens(h.goroutinesCtx)
		}
	}
}

func (h *Handler) updateChats() {
	defer h.wg.Done()

	ticker := time.NewTicker(updateChatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		case <-ticker.C:
			chats, err := h.useCases.GetUserChats(h.goroutinesCtx, nil)
			if err != nil {
				logging.LogErrorContext(
					h.goroutinesCtx,
					h.logger,
					"Не удалось обновить список чатов",
					err,
				)
			}

			// Отправляем событие на фронтенд
			runtime.EventsEmit(h.wailsCtx, chatsUpdatedEventName, chats)
		}
	}
}
