# Пакет internal/v2/handlers/messages

## Назначение

Хендлер сообщений. Управляет получением, отправкой и удалением сообщений в чатах. Поддерживает ответ на сообщение.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер сообщений. Содержит `useCases` и `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetChatMessages(chatID uint64, pagination *domains.Pagination) ([]domains.Message, error)` | Возвращает сообщения чата с пагинацией. |
| `(h *Handler) SendMessage(chatID uint64, text string, replyToMessageID *uint64) error` | Отправка сообщения в чат. Автоматически определяет отправителя через `GetCurrentUser`. Поддерживает ответ на сообщение (`replyToMessageID`). |
| `(h *Handler) DeleteMessage(messageID uint64, forAll bool) error` | Удаление сообщения (для себя или для всех). |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `Message`, `DeleteMessageDTO`, `Pagination`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`
