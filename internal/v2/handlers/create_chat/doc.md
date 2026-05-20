# Пакет internal/v2/handlers/create_chat

## Назначение

Хендлер создания чатов. Валидирует входные данные, формирует доменный объект `Chat` и делегирует создание use cases.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер создания чатов. Содержит `useCases`, `errorsMapper`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) CreateChat(in domains.CreateChatDTO) (*domains.Chat, error)` | Создание чата. Валидирует DTO через `IsValid()`, формирует список участников из `MemberIDs` и вызывает `useCases.CreateChat`. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `CreateChatDTO`, `Chat`, `User`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrInvalidChat`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
