package theme

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

func (h *Handler) GetTheme() domains.ThemeType {
	ctx := context.Background()

	return h.useCases.GetTheme(ctx)
}

func (h *Handler) ToggleTheme() (domains.ThemeType, error) {
	ctx := context.Background()

	currentTheme := h.useCases.GetTheme(ctx)

	newTheme := domains.ThemeLight
	if currentTheme == domains.ThemeLight {
		newTheme = domains.ThemeDark
	}

	if err := h.useCases.SetTheme(ctx, newTheme); err != nil {
		return domains.ThemeLight, err
	}

	return newTheme, nil
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
