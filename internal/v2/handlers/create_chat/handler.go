package create_chat

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type Handler struct {
	useCases interfaces.UseCases

	ctx context.Context
}

func New(useCases interfaces.UseCases) *Handler {
	return &Handler{
		useCases: useCases,
		ctx:      context.Background(),
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

type CreateChatRequest struct {
	Type    string   `json:"type"`
	Members []uint64 `json:"members"`
	Title   *string  `json:"title,omitempty"`
}

func (h *Handler) CreateChat(
	req CreateChatRequest,
) (*domains.Chat, error) {
	members := make([]domains.User, len(req.Members))
	for i, id := range req.Members {
		members[i] = domains.User{ID: id}
	}

	chat := &domains.Chat{
		Type:    domains.ChatType(req.Type),
		Members: members,
		Title:   req.Title,
	}

	return h.useCases.CreateChat(h.ctx, *chat)
}
