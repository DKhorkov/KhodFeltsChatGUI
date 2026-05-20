# Пакет internal/v2/handlers/notification

## Назначение

Хендлер системных уведомлений (desktop notifications). Инициализирует подсистему уведомлений Wails, запрашивает разрешения, отправляет уведомления и обрабатывает клик по уведомлению (открытие чата).

## Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `openChatEventName` | `"open_chat"` | Имя Wails-события открытия чата при клике на уведомление. |

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер уведомлений. Содержит `logger`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(logger logging.Logger) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. Инициализирует уведомления, запрашивает разрешение и регистрирует обработчик кликов по уведомлениям. |
| `(h *Handler) ShowNotification(title, body string, chatID int) error` | Отправляет системное уведомление с привязкой к чату (`chatId` в `Data`). |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

### Приватные методы

| Сигнатура | Описание |
|-----------|----------|
| `(h *Handler) focusWindow()` | Восстанавливает и выводит окно на передний план (unminimise, show, always on top). |

## Зависимости

- `github.com/DKhorkov/libs/logging` — логирование
- `github.com/wailsapp/wails/v2/pkg/runtime` — `InitializeNotifications`, `RequestNotificationAuthorization`, `OnNotificationResponse`, `SendNotification`, `EventsEmit`, управление окном
