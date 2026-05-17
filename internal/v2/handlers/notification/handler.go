package notification

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/libs/logging"
	"github.com/gen2brain/beeep"
)

type Handler struct {
	logger   logging.Logger
	wailsCtx context.Context
}

func New(logger logging.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) ShowNotification(title, body string) error {
	if err := beeep.Notify(title, body, common.AppIconPath()); err != nil {
		logging.LogError(h.logger, "Ошибка отправки системного уведомления", err)

		return err
	}

	return nil
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
