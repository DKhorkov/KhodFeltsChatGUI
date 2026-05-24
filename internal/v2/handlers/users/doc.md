# Пакет internal/v2/handlers/users

## Назначение

Хендлер пользователей. Управляет получением текущего пользователя, поиском пользователей и обновлением профиля. Выполняет валидацию username перед обновлением.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер пользователей. Содержит `useCases`, `errorsMapper`, `validationConfig`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, validationConfig config.ValidationConfig) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetCurrentUser() (*domains.User, error)` | Возвращает текущего пользователя. |
| `(h *Handler) SearchUsers(filters *domains.UsersFilters, pagination *domains.Pagination) ([]domains.User, error)` | Поиск пользователей по фильтрам с пагинацией. |
| `(h *Handler) UpdateUser(in domains.UpdateUserDTO) (*domains.User, error)` | Обновление профиля. Валидирует username (если задан) по `UsernameRegExps`. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/config` — `ValidationConfig`
- `github.com/DKhorkov/kfcGUI/internal/domains` — `User`, `UsersFilters`, `UpdateUserDTO`, `Pagination`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrInvalidLogin`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/validation` — валидация по регулярным выражениям
