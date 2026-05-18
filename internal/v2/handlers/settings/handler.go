package settings

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type Handler struct {
	useCases interfaces.UseCases

	wailsCtx context.Context
}

func New(
	useCases interfaces.UseCases,
) *Handler {
	return &Handler{
		useCases: useCases,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) GetSettings() (*domains.Settings, error) {
	ctx := context.Background()

	return h.useCases.GetSettings(ctx)
}

func (h *Handler) UpdateSettings(settings domains.Settings) (*domains.Settings, error) {
	ctx := context.Background()

	return h.useCases.UpdateSettings(ctx, settings)
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
