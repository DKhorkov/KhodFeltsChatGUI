# Пакет internal/v2/handlers/chats

## Назначение

Хендлер чатов. Управляет получением и созданием чатов. Запускает фоновые горутины для чтения WebSocket-событий, обновления токенов и периодического обновления списка чатов.

## Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `refreshTokensInterval` | `1 * time.Minute` | Интервал обновления токенов. |
| `updateChatsInterval` | `5 * time.Second` | Интервал обновления списка чатов. |
| `chatsUpdatedEventName` | `"chats_updated"` | Имя Wails-события обновления чатов. |
| `newMessageEventName` | `"new_message"` | Имя Wails-события нового сообщения. |
| `messageDeletedEventName` | `"message_deleted"` | Имя Wails-события удаления сообщения. |

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер чатов. Содержит `useCases`, `errorsMapper`, `logger`, контексты горутин и `sync.WaitGroup`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, logger logging.Logger) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error)` | Возвращает чаты пользователя с пагинацией. |
| `(h *Handler) CreateChat(in domains.CreateChatDTO) (*domains.Chat, error)` | Создание чата. Валидирует DTO через `IsValid()`, формирует список участников и вызывает use case. |
| `(h *Handler) StartListening()` | Запускает фоновые горутины: чтение WS-событий, обновление токенов, обновление чатов. |
| `(h *Handler) StopListening()` | Останавливает все фоновые горутины и ожидает их завершения. |

### Приватные методы

| Сигнатура | Описание |
|-----------|----------|
| `(h *Handler) readEvents()` | Горутина чтения WS-событий. Диспатчит по типу: `new_message` → эмит `new_message`, `message_deleted` → эмит `message_deleted`. |
| `(h *Handler) refreshTokens()` | Горутина периодического обновления токенов (раз в минуту). |
| `(h *Handler) updateChats()` | Горутина периодического обновления списка чатов (раз в 5 секунд). Эмитит событие `chats_updated`. |

## Тесты (`handler_test.go`)

| Тест | Описание |
|------|----------|
| `TestHandler_GetUserChats` | Получение чатов: без пагинации, с пагинацией, пустой список, ошибка use case. |
| `TestHandler_CreateChat` | Создание чата: приватный, групповой, невалидный DTO, ошибка use case. |
| `TestHandler_SetContext` | Установка Wails-контекста. |
| `TestHandler_StartListening_StopListening` | Запуск/остановка горутин: выход по `ErrWebsocketClosed`, остановка без запуска. |
| `TestHandler_ReadEvents` | Горутина чтения WS-событий: generic ошибка + продолжение, невалидный JSON для `new_message`, невалидный JSON для `message_deleted`, неизвестный тип события, выход по отмене контекста. |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `Chat`, `CreateChatDTO`, `Pagination`, `WSEvent`, `MessageDeletedPayload`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrWebsocketClosed`, `ErrInvalidChat`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/logging` — логирование
- `github.com/wailsapp/wails/v2/pkg/runtime` — `EventsEmit`
