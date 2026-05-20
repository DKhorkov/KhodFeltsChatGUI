# Пакет internal/v2/handlers/chat

## Назначение

Основной хендлер чата. Управляет получением чатов, сообщений, отправкой сообщений, переключением темы. Запускает фоновые горутины для чтения WebSocket-сообщений, обновления токенов и периодического обновления списка чатов.

## Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `refreshTokensInterval` | `1 * time.Minute` | Интервал обновления токенов. |
| `updateChatsInterval` | `5 * time.Second` | Интервал обновления списка чатов. |
| `chatsUpdatedEventName` | `"chats_updated"` | Имя Wails-события обновления чатов. |
| `newMessageEventName` | `"new_message"` | Имя Wails-события нового сообщения. |

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер чата. Содержит `useCases`, `errorsMapper`, `logger`, `validationConfig`, контексты горутин и `sync.WaitGroup`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, logger logging.Logger, validationConfig config.ValidationConfig) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetCurrentUser() (*domains.User, error)` | Возвращает текущего пользователя. |
| `(h *Handler) GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error)` | Возвращает чаты пользователя с пагинацией. |
| `(h *Handler) GetChatMessages(chatID uint64, pagination *domains.Pagination) ([]domains.Message, error)` | Возвращает сообщения чата с пагинацией. |
| `(h *Handler) SendMessage(chatID uint64, text string) error` | Отправка сообщения в чат. Автоматически определяет отправителя. |
| `(h *Handler) ToggleTheme() (domains.ThemeType, error)` | Переключает тему (светлая/тёмная). |
| `(h *Handler) StartListening()` | Запускает фоновые горутины: чтение сообщений, обновление токенов, обновление чатов. |
| `(h *Handler) StopListening()` | Останавливает все фоновые горутины и ожидает их завершения. |

### Приватные методы

| Сигнатура | Описание |
|-----------|----------|
| `(h *Handler) readMessages()` | Горутина чтения сообщений из WebSocket. Эмитит событие `new_message`. |
| `(h *Handler) refreshTokens()` | Горутина периодического обновления токенов (раз в минуту). |
| `(h *Handler) updateChats()` | Горутина периодического обновления списка чатов (раз в 5 секунд). Эмитит событие `chats_updated`. |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/common` — `Timezone`
- `github.com/DKhorkov/kfcGUI/internal/config` — `ValidationConfig`
- `github.com/DKhorkov/kfcGUI/internal/domains` — `User`, `Chat`, `Message`, `Pagination`, `ThemeType`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrWebsocketClosed`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/logging` — логирование
- `github.com/wailsapp/wails/v2/pkg/runtime` — `EventsEmit`
