package chat

import (
	"context"
	"sync"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Handler struct {
	useCases         interfaces.UseCases
	errorsMapper     interfaces.ErrorsMapper
	validationConfig config.ValidationConfig

	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	currentUser *domains.User
}

func New(
	useCases interfaces.UseCases,
	errorsMapper interfaces.ErrorsMapper,
	validationConfig config.ValidationConfig,
) *Handler {
	return &Handler{
		useCases:         useCases,
		errorsMapper:     errorsMapper,
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

func (h *Handler) GetUserChats(limit, offset int) ([]domains.Chat, error) {
	return h.useCases.GetUserChats(h.ctx, limit, offset)
}

func (h *Handler) GetChatMessages(
	chatID uint64,
	limit, offset int,
) ([]domains.Message, error) {
	return h.useCases.GetChatMessages(h.ctx, chatID, limit, offset)
}

type SendMessageRequest struct {
	ChatID  uint64 `json:"chatId"`
	Message string `json:"message"`
}

func (h *Handler) SendMessage(req SendMessageRequest) error {
	sender, err := h.useCases.GetCurrentUser(h.ctx)
	if err != nil {
		return err
	}

	message := domains.Message{
		ChatID: req.ChatID,
		Text:   req.Message,
		Sender: domains.User{
			ID: sender.ID,
		},
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
				time.Sleep(1 * time.Second)
				continue
			}

			// Отправляем событие на фронтенд
			runtime.EventsEmit(h.ctx, "new_message", message)
		}
	}
}

func (h *Handler) refreshTokens() {
	defer h.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.useCases.RefreshTokens(h.ctx)
		}
	}
}

func (h *Handler) updateChats() {
	defer h.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			chats, err := h.useCases.GetUserChats(h.ctx, 0, 0)
			if err == nil {
				runtime.EventsEmit(h.ctx, "chats_updated", chats)
			}
		}
	}
}
