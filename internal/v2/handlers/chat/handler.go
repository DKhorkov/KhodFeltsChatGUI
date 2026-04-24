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
)

type Handler struct {
	useCases         interfaces.UseCases
	errorsMapper     interfaces.ErrorsMapper
	logger           logging.Logger
	validationConfig config.ValidationConfig

	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
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
		ctx:              context.Background(),
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *Handler) GetCurrentUser() (*domains.User, error) {
	return h.useCases.GetCurrentUser(h.ctx)
}

func (h *Handler) GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error) {
	return h.useCases.GetUserChats(h.ctx, pagination)
}

func (h *Handler) GetChatMessages(
	chatID uint64,
	pagination *domains.Pagination,
) ([]domains.Message, error) {
	return h.useCases.GetChatMessages(h.ctx, chatID, pagination)
}

func (h *Handler) SendMessage(chatID uint64, text string) error {
	sender, err := h.useCases.GetCurrentUser(h.ctx)
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

	return h.useCases.SendMessage(h.ctx, message)
}

func (h *Handler) StartListening() {
	h.ctx, h.cancelFunc = context.WithCancel(h.ctx)

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
	if h.cancelFunc != nil {
		h.cancelFunc()
	}

	h.wg.Wait()
}

func (h *Handler) readMessages() {
	defer h.wg.Done()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			message, err := h.useCases.ReadMessage(h.ctx)
			if err != nil {
				// Соккет закрыт, отключаем горутину
				if errors.Is(err, customerrors.ErrWebsocketClosed) {
					logging.LogInfo(
						h.logger,
						"startReadMessagesGoroutine завершена из-за закрытия соединения",
					)

					return
				}

				logging.LogErrorContext(h.ctx, h.logger, "Не удалось прочитать сообщение", err)

				continue
			}

			// Отправляем событие на фронтенд
			runtime.EventsEmit(h.ctx, "new_message", message)
		}
	}
}

func (h *Handler) refreshTokens() {
	defer h.wg.Done()

	ticker := time.NewTicker(refreshTokensInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			_, _ = h.useCases.RefreshTokens(h.ctx)
		}
	}
}

func (h *Handler) updateChats() {
	defer h.wg.Done()

	ticker := time.NewTicker(updateChatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			chats, err := h.useCases.GetUserChats(h.ctx, nil)
			if err == nil {
				runtime.EventsEmit(h.ctx, "chats_updated", chats)
			}
		}
	}
}
