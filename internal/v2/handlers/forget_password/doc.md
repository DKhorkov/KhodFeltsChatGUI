# Пакет internal/v2/handlers/forget_password

## Назначение

Хендлер сброса пароля по токену. Валидирует токен и новый пароль, делегирует выполнение use cases.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер сброса пароля. Содержит `useCases`, `errorsMapper`, `validationConfig`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, validationConfig config.ValidationConfig) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) ForgetPassword(forgetPasswordToken string, in domains.ForgetPasswordDTO) error` | Сброс пароля. Проверяет непустоту токена, валидирует новый пароль и вызывает `useCases.ForgetPassword`. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/config` — `ValidationConfig`
- `github.com/DKhorkov/kfcGUI/internal/domains` — `ForgetPasswordDTO`
- `github.com/DKhorkov/kfcGUI/internal/errors` — `ErrInvalidForgetPasswordToken`, `ErrInvalidPassword`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/validation` — валидация по регулярным выражениям
