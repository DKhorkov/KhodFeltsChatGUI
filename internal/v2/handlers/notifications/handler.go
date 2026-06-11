package notifications

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/logging"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	openChatEventName = "open_chat"

	maxBadgeNumber = 99
)

type Handler struct {
	logger   logging.Logger
	appTitle string
	wailsCtx context.Context
}

func New(logger logging.Logger, appTitle string) *Handler {
	return &Handler{
		logger:   logger,
		appTitle: appTitle,
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

// SetUnreadBadge обновляет заголовок окна, добавляя префикс с числом непрочитанных
// сообщений. При total <= 0 заголовок сбрасывается на исходный. Числа > 99
// отображаются как "99+".
func (h *Handler) SetUnreadBadge(total int) {
	wailsruntime.WindowSetTitle(h.wailsCtx, formatBadgeTitle(total, h.appTitle))
}

func formatBadgeTitle(total int, appTitle string) string {
	switch {
	case total <= 0:
		return appTitle
	case total > maxBadgeNumber:
		return fmt.Sprintf("(%d+) %s", maxBadgeNumber, appTitle)
	default:
		return fmt.Sprintf("(%d) %s", total, appTitle)
	}
}

func (h *Handler) StartListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) StopListening() {} //nolint:revive // Удалится в будущем при добавлении функционала

func (h *Handler) focusWindow() {
	wailsruntime.WindowUnminimise(h.wailsCtx)
	wailsruntime.WindowShow(h.wailsCtx)
	wailsruntime.WindowSetAlwaysOnTop(h.wailsCtx, true)
	wailsruntime.WindowSetAlwaysOnTop(h.wailsCtx, false)
}
