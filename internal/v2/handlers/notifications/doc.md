# Пакет internal/v2/handlers/notification

## Назначение

Хендлер системных уведомлений (desktop notifications). Инициализирует подсистему уведомлений Wails, запрашивает разрешения, отправляет уведомления, обрабатывает клик по уведомлению (открытие чата) и обновляет счётчик непрочитанных сообщений в заголовке окна.

## Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `openChatEventName` | `"open_chat"` | Имя Wails-события открытия чата при клике на уведомление. |
| `maxBadgeNumber` | `99` | Максимальное число, отображаемое в заголовке окна. Большие значения выводятся как `(99+)`. |

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер уведомлений. Содержит `logger`, `appTitle` (базовый заголовок окна, используется при сборке префикса со счётчиком), `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(logger logging.Logger, appTitle string) *Handler` | Конструктор хендлера. `appTitle` должен совпадать с `options.App.Title` из `cmd/v2/main.go`. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. Инициализирует уведомления, запрашивает разрешение и регистрирует обработчик кликов по уведомлениям. |
| `(h *Handler) ShowNotification(title, body string, chatID int) error` | Отправляет системное уведомление с привязкой к чату (`chatId` в `Data`). |
| `(h *Handler) SetUnreadBadge(total int)` | Обновляет заголовок окна префиксом `(N)` со счётчиком непрочитанных сообщений. При `total <= 0` заголовок сбрасывается на `appTitle`. Числа > 99 отображаются как `(99+)`. Биндится во фронт как `wailsjs/go/notifications/Handler#SetUnreadBadge`. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

### Приватные методы

| Сигнатура | Описание |
|-----------|----------|
| `(h *Handler) focusWindow()` | Восстанавливает и выводит окно на передний план (unminimise, show, always on top). |
| `formatBadgeTitle(total int, appTitle string) string` | Чистая функция форматирования заголовка с префиксом. Покрыта unit-тестами. |

## Зависимости

- `github.com/DKhorkov/libs/logging` — логирование
- `github.com/wailsapp/wails/v2/pkg/runtime` — `InitializeNotifications`, `RequestNotificationAuthorization`, `OnNotificationResponse`, `SendNotification`, `EventsEmit`, `WindowSetTitle`, управление окном
