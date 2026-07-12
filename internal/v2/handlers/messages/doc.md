# Пакет internal/v2/handlers/messages

## Назначение

Хендлер сообщений. Управляет получением, отправкой, удалением и редактированием сообщений в чатах. Поддерживает ответ на сообщение.

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
| `(h *Handler) GetMessageByID(messageID uint64) (*domains.Message, error)` | Получение сообщения по ID. |
| `(h *Handler) DeleteMessage(messageID uint64, forAll bool) error` | Удаление сообщения (для себя или для всех). |
| `(h *Handler) UpdateMessage(messageID uint64, text string) error` | Редактирование сообщения. Создаёт `UpdateMessageDTO` и вызывает usecases. |
| `(h *Handler) ListReactions() ([]domains.Reaction, error)` | Возвращает справочник emoji-реакций для UI-пикера. Вызывается Vue-фронтом через Wails-биндинг. |
| `(h *Handler) AddMessageReaction(messageID, reactionID uint64) error` | Ставит реакцию на сообщение. Ошибка `already exists` (409 от API) означает, что реакция уже стоит — Vue-компонент по этой ошибке делает автоматический toggle через `RemoveMessageReaction`. |
| `(h *Handler) RemoveMessageReaction(messageID, reactionID uint64) error` | Снимает реакцию. Сервер идемпотентен → nil при 204. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `Message`, `DeleteMessageDTO`, `UpdateMessageDTO`, `Pagination`, `Reaction`, `MessageReactionDTO`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`
