# Пакет internal/v2/handlers/search_users

## Назначение

Хендлер поиска пользователей. Делегирует поиск с фильтрами и пагинацией в use cases.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер поиска пользователей. Содержит `useCases`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) SearchUsers(filters *domains.UsersFilters, pagination *domains.Pagination) ([]domains.User, error)` | Поиск пользователей по фильтрам с пагинацией. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `UsersFilters`, `Pagination`, `User`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`
