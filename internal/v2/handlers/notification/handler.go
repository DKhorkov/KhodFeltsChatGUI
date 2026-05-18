package notification

import (
	"context"

	"github.com/DKhorkov/libs/logging"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const openChatEventName = "open_chat"

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

	if err := wailsruntime.InitializeNotifications(ctx); err != nil {
		logging.LogError(h.logger, "Ошибка инициализации уведомлений", err)
		return
	}

	authorized, err := wailsruntime.RequestNotificationAuthorization(ctx)
	if err != nil {
		logging.LogError(h.logger, "Ошибка запроса разрешения на уведомления", err)
	} else {
		logging.LogInfo(h.logger, "Разрешение на уведомления", "authorized", authorized)
	}

	wailsruntime.OnNotificationResponse(ctx, func(result wailsruntime.NotificationResult) {
		if result.Error != nil {
			logging.LogError(h.logger, "Ошибка ответа на уведомление", result.Error)
			return
		}

		chatID, ok := result.Response.UserInfo["chatId"]
		if !ok {
			return
		}

		h.focusWindow()
		wailsruntime.EventsEmit(h.wailsCtx, openChatEventName, chatID)
	})
}

func (h *Handler) ShowNotification(title, body string, chatID int) error {
	if err := wailsruntime.SendNotification(h.wailsCtx, wailsruntime.NotificationOptions{
		Title: title,
		Body:  body,
		Data: map[string]any{
			"chatId": chatID,
		},
	}); err != nil {
		logging.LogError(h.logger, "Ошибка отправки системного уведомления", err)

		return err
	}

	return nil
}

func (h *Handler) focusWindow() {
	wailsruntime.WindowUnminimise(h.wailsCtx)
	wailsruntime.WindowShow(h.wailsCtx)
	wailsruntime.WindowSetAlwaysOnTop(h.wailsCtx, true)
	wailsruntime.WindowSetAlwaysOnTop(h.wailsCtx, false)
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала
