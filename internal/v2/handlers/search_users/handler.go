package search_users

import (
	"context"

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

func (h *Handler) SearchUsers(
	filters *domains.UsersFilters,
	pagination *domains.Pagination,
) ([]domains.User, error) {
	ctx := context.Background()

	return h.useCases.SearchUsers(ctx, filters, pagination)
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
