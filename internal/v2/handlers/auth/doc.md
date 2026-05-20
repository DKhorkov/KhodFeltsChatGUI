# Пакет internal/v2/handlers/auth

## Назначение

Хендлер аутентификации и регистрации пользователей. Выполняет валидацию входных данных (email, username, пароль) перед вызовом use cases.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер аутентификации. Содержит `useCases`, `errorsMapper`, `validationConfig`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases, errorsMapper interfaces.ErrorsMapper, validationConfig config.ValidationConfig) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) Login(in domains.LoginDTO) error` | Авторизация. Валидирует логин (email или username) и пароль. |
| `(h *Handler) Register(in domains.RegisterDTO) error` | Регистрация. Валидирует email, username и пароль. |
| `(h *Handler) SendVerifyEmail(email string) error` | Отправка письма для подтверждения email. Валидирует email. |
| `(h *Handler) SendForgetPassword(email string) error` | Отправка письма для сброса пароля. Валидирует email. |
| `(h *Handler) ForgetPassword(token string, in domains.ForgetPasswordDTO) error` | Сброс пароля по токену. Валидирует новый пароль. |
| `(h *Handler) Authenticate() error` | Проверка текущей аутентификации пользователя. |
| `(h *Handler) Logout() error` | Выход из системы. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/config` — `ValidationConfig`
- `github.com/DKhorkov/kfcGUI/internal/domains` — `LoginDTO`, `RegisterDTO`, `ForgetPasswordDTO`
- `github.com/DKhorkov/kfcGUI/internal/errors` — кастомные ошибки валидации
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/validation` — валидация по регулярным выражениям
