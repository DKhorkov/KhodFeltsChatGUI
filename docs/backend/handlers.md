# Wails-хендлеры

Пакет: `internal/v2/handlers/`

## Общий паттерн

Каждый хендлер:
- реализует интерфейс `interfaces.Handler` (`SetContext`, `StartListening`, `StopListening`);
- создаётся функцией `New(...)`, принимающей зависимости;
- хранит `wailsCtx context.Context`, устанавливаемый при вызове `Startup`;
- для запросов к UseCase всегда создаёт `context.Background()` (не использует `wailsCtx` напрямую в вызовах, кроме эмита событий);
- регистрируется в Wails через `app.BindHandlers()`, после чего Wails генерирует TypeScript-обёртки для всех публичных методов.

---

## auth.Handler

Файл: `internal/v2/handlers/auth/handler.go`

### Конструктор

```go
func New(
    useCases interfaces.UseCases,
    errorsMapper interfaces.ErrorsMapper,
    validationConfig config.ValidationConfig,
) *Handler
```

### Методы

#### `Login(in domains.LoginDTO) error`
Выполняет вход пользователя. Перед вызовом UseCase проверяет:
- `in.Login` соответствует регулярному выражению email **или** правилам username;
- `in.Password` соответствует всем правилам пароля.

#### `Register(in domains.RegisterDTO) error`
Регистрирует нового пользователя. Валидирует email, username и пароль. При нарушении возвращает смапированную ошибку.

#### `SendVerifyEmail(email string) error`
Отправляет письмо для подтверждения email. Валидирует формат email.

#### `SendForgetPassword(email string) error`
Отправляет письмо со ссылкой для сброса пароля. Валидирует формат email.

#### `ForgetPassword(token string, in domains.ForgetPasswordDTO) error`
Применяет новый пароль по токену из письма. Валидирует новый пароль.

#### `Authenticate() error`
Проверяет, аутентифицирован ли пользователь (обновляет токены и получает текущего пользователя). Используется при запуске приложения для восстановления сессии.

#### `Logout() error`
Завершает сессию пользователя.

### Фоновые горутины
Нет. `StartListening` и `StopListening` — пустые методы (заглушки для будущего функционала).

---

## chat.Handler

Файл: `internal/v2/handlers/chat/handler.go`

### Конструктор

```go
func New(
    useCases interfaces.UseCases,
    errorsMapper interfaces.ErrorsMapper,
    logger logging.Logger,
    validationConfig config.ValidationConfig,
) *Handler
```

### Методы

#### `GetCurrentUser() (*domains.User, error)`
Возвращает текущего аутентифицированного пользователя (с обновлением токенов).

#### `GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error)`
Возвращает список чатов текущего пользователя с поддержкой пагинации.

#### `GetChatMessages(chatID uint64, pagination *domains.Pagination) ([]domains.Message, error)`
Возвращает историю сообщений указанного чата. Время сообщений приводится к локальному часовому поясу пользователя.

#### `SendMessage(chatID uint64, text string) error`
Отправляет сообщение в чат через WebSocket. Внутри получает текущего пользователя, формирует `domains.Message` и передаёт в UseCase.

#### `ToggleTheme() (domains.ThemeType, error)`
Переключает тему между светлой и тёмной и сохраняет выбор.

### Фоновые горутины

`StartListening()` запускает три горутины:

| Горутина | Интервал | Описание |
|---|---|---|
| `readMessages` | постоянно | Читает входящие сообщения из WebSocket и эмитит событие `"new_message"` во фронтенд через `runtime.EventsEmit` |
| `refreshTokens` | 1 минута | Периодически обновляет JWT-токены |
| `updateChats` | 5 секунд | Периодически получает список чатов и эмитит событие `"chats_updated"` во фронтенд |

`StopListening()` отменяет контекст горутин и ожидает их завершения через `sync.WaitGroup`.

---

## create_chat.Handler

Файл: `internal/v2/handlers/create_chat/handler.go`

### Конструктор

```go
func New(
    useCases interfaces.UseCases,
    errorsMapper interfaces.ErrorsMapper,
) *Handler
```

### Методы

#### `CreateChat(in domains.CreateChatDTO) (*domains.Chat, error)`
Создаёт новый чат. Перед вызовом UseCase:
- вызывает `in.IsValid()` для проверки типа чата и количества участников;
- преобразует `MemberIDs` в срез `[]domains.User{ID: id}`.

### Фоновые горутины
Нет.

---

## search_users.Handler

Файл: `internal/v2/handlers/search_users/handler.go`

### Конструктор

```go
func New(useCases interfaces.UseCases) *Handler
```

Не принимает `errorsMapper` — ошибки возвращаются как есть из UseCase (уже смапированы).

### Методы

#### `SearchUsers(filters *domains.UsersFilters, pagination *domains.Pagination) ([]domains.User, error)`
Выполняет поиск пользователей по фильтрам. Текущий пользователь исключается из результатов на уровне UseCase.

### Фоновые горутины
Нет.

---

## forget_password.Handler

Файл: `internal/v2/handlers/forget_password/handler.go`

### Конструктор

```go
func New(
    useCases interfaces.UseCases,
    errorsMapper interfaces.ErrorsMapper,
    validationConfig config.ValidationConfig,
) *Handler
```

### Методы

#### `ForgetPassword(forgetPasswordToken string, in domains.ForgetPasswordDTO) error`
Сбрасывает пароль по токену. Отличается от `auth.Handler.ForgetPassword` тем, что дополнительно проверяет непустоту токена перед валидацией пароля.

### Фоновые горутины
Нет.

---

## profile.Handler

Файл: `internal/v2/handlers/profile/handler.go`

### Конструктор

```go
func New(
    useCases interfaces.UseCases,
    errorsMapper interfaces.ErrorsMapper,
    validationConfig config.ValidationConfig,
) *Handler
```

### Методы

#### `ChangePassword(in domains.ChangePasswordDTO) error`
Меняет пароль аутентифицированного пользователя. Валидирует оба пароля (старый и новый) по правилам пароля.

#### `UpdateUser(in domains.UpdateUserDTO) (*domains.User, error)`
Обновляет профиль пользователя. Если `in.Username != nil`, валидирует новый логин по правилам username.

### Фоновые горутины
Нет.

---

## theme.Handler

Файл: `internal/v2/handlers/theme/handler.go`

### Конструктор

```go
func New(useCases interfaces.UseCases) *Handler
```

Не принимает `errorsMapper` и `validationConfig`.

### Методы

#### `GetTheme() domains.ThemeType`
Возвращает текущую тему из локальных настроек. При ошибке чтения возвращает `ThemeLight`.

#### `ToggleTheme() (domains.ThemeType, error)`
Переключает тему и сохраняет. Аналогичен методу `chat.Handler.ToggleTheme`.

### Фоновые горутины
Нет.
