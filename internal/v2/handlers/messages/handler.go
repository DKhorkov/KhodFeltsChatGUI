package messages

import (
	"context"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type Handler struct {
	useCases interfaces.UseCases

	wailsCtx context.Context
}

func New(useCases interfaces.UseCases) *Handler {
	return &Handler{
		useCases: useCases,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) GetChatMessages(
	chatID uint64,
	pagination *domains.Pagination,
) ([]domains.Message, error) {
	ctx := context.Background()

	return h.useCases.GetChatMessages(ctx, chatID, pagination)
}

func (h *Handler) SendMessage(chatID uint64, text string, replyToMessageID *uint64) error {
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

	if replyToMessageID != nil {
		message.ReplyToMessage = &domains.Message{ID: *replyToMessageID}
	}

	return h.useCases.SendMessage(ctx, message)
}

func (h *Handler) GetMessageByID(messageID uint64) (*domains.Message, error) {
	ctx := context.Background()

	return h.useCases.GetMessageByID(ctx, messageID)
}

func (h *Handler) UpdateMessage(messageID uint64, text string) error {
	ctx := context.Background()

	return h.useCases.UpdateMessage(ctx, domains.UpdateMessageDTO{
		MessageID: messageID,
		Text:      text,
	})
}

func (h *Handler) DeleteMessage(messageID uint64, forAll bool) error {
	ctx := context.Background()

	return h.useCases.DeleteMessage(ctx, domains.DeleteMessageDTO{
		MessageID: messageID,
		ForAll:    forAll,
	})
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
