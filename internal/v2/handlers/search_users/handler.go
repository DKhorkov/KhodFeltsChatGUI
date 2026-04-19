package search_users

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

func (h *Handler) SearchUsers(
	username string,
	limit, offset int,
) ([]domains.User, error) {
	return h.useCases.SearchUsers(context.Background(), username, limit, offset)
}
