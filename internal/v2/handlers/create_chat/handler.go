package create_chat

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type Handler struct {
	useCases interfaces.UseCases
}

func New(useCases interfaces.UseCases) *Handler {
	return &Handler{
		useCases: useCases,
	}
}

type CreateChatRequest struct {
	Type    string   `json:"type"`
	Members []uint64 `json:"members"`
	Title   *string  `json:"title,omitempty"`
}

func (h *Handler) SearchUsers(
	ctx context.Context,
	username string,
	limit, offset int,
) ([]domains.User, error) {
	return h.useCases.SearchUsers(ctx, username, limit, offset)
}

func (h *Handler) CreateChat(
	ctx context.Context,
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

	return h.useCases.CreateChat(ctx, *chat)
}
