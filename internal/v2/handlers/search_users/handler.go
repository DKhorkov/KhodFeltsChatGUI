package search_users

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

func (h *Handler) SearchUsers(
	filters *domains.UsersFilters,
	pagination *domains.Pagination,
) ([]domains.User, error) {
	return h.useCases.SearchUsers(h.ctx, filters, pagination)
}
