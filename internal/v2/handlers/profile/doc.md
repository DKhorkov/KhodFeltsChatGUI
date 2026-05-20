# Пакет internal/v2/handlers/profile

## Назначение

Хендлер профиля пользователя. Обрабатывает смену пароля и обновление данных пользователя с валидацией.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер профиля. Содержит `useCases`, `errorsMapper`, `validationConfig`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, validationConfig config.ValidationConfig) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) ChangePassword(in domains.ChangePasswordDTO) error` | Смена пароля. Валидирует и новый, и старый пароль. |
| `(h *Handler) UpdateUser(in domains.UpdateUserDTO) (*domains.User, error)` | Обновление данных пользователя. Валидирует username (если передан). Возвращает обновлённого пользователя. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/config` — `ValidationConfig`
- `github.com/DKhorkov/kfcGUI/internal/domains` — `ChangePasswordDTO`, `UpdateUserDTO`, `User`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrInvalidPassword`, `ErrInvalidLogin`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/validation` — валидация по регулярным выражениям
