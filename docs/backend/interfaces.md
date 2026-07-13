# Интерфейсы

Пакет: `internal/interfaces/`

Все интерфейсы снабжены директивами `//go:generate mockgen ...` для автоматической генерации моков в пакете `mocks/`.

---

## Application

Файл: `internal/interfaces/application.go`

```go
type Application interface {
    Startup(ctx context.Context)
    Shutdown(ctx context.Context)
    BindHandlers() []any
}
```

Реализован структурой `internal/v2/application.App`. Используется Wails для управления жизненным циклом приложения.

---

## Handler

Файл: `internal/interfaces/handler.go`

```go
type Handler interface {
    SetContext(ctx context.Context)
    StartListening()
    StopListening()
}
```

Реализован всеми хендлерами в `internal/v2/handlers/`. `SetContext` вызывается при `Startup`, передавая Wails-контекст (необходим для `runtime.EventsEmit`). `StartListening` / `StopListening` управляют фоновыми горутинами хендлера.

---

## UseCases

Файл: `internal/interfaces/usecases.go`

```go
type UseCases interface {
    // Аутентификация
    Authenticate(ctx context.Context) (*domains.User, error)
    GetCurrentUser(ctx context.Context) (*domains.User, error)
    Login(ctx context.Context, in domains.LoginDTO) (*domains.User, error)
    Logout(ctx context.Context) error
    Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
    RefreshTokens(ctx context.Context) (*domains.TokensDTO, error)
    SendVerifyEmailMessage(ctx context.Context, email string) error
    SendForgetPasswordMessage(ctx context.Context, email string) error
    ForgetPassword(ctx context.Context, forgetPasswordToken, newPassword string) error
    ChangePassword(ctx context.Context, changePasswordData domains.ChangePasswordDTO) error

    // Сообщения
    SendMessage(ctx context.Context, message domains.Message) error
    ReadMessage(ctx context.Context) (*domains.Message, error)
    GetChatMessages(ctx context.Context, chatID uint64, pagination *domains.Pagination) ([]domains.Message, error)

    // Реакции
    ListReactions(ctx context.Context) ([]domains.Reaction, error)
    AddMessageReaction(ctx context.Context, dto domains.MessageReactionDTO) error
    RemoveMessageReaction(ctx context.Context, dto domains.MessageReactionDTO) error

    // Чаты
    CreateChat(ctx context.Context, chat domains.Chat) (*domains.Chat, error)
    GetUserChats(ctx context.Context, pagination *domains.Pagination) ([]domains.Chat, error)

    // Пользователи
    UpdateUser(ctx context.Context, updateUserData domains.UpdateUserDTO) (*domains.User, error)
    SearchUsers(ctx context.Context, filters *domains.UsersFilters, pagination *domains.Pagination) ([]domains.User, error)

    // Настройки
    GetTheme(ctx context.Context) domains.ThemeType
    SetTheme(ctx context.Context, theme domains.ThemeType) error
}
```

Реализован структурой `internal/usecases.UseCases`.

---

## AuthRepository

Файл: `internal/interfaces/repositories.go`

```go
type AuthRepository interface {
    Register(ctx context.Context, registerData domains.RegisterDTO) (*domains.User, error)
    Login(ctx context.Context, in domains.LoginDTO) (*domains.TokensDTO, error)
    Logout(ctx context.Context, accessToken string) error
    RefreshTokens(ctx context.Context, refreshToken string) (*domains.TokensDTO, error)
    SendVerifyEmailMessage(ctx context.Context, email string) error
    SendForgetPasswordMessage(ctx context.Context, email string) error
    ForgetPassword(ctx context.Context, forgetPasswordToken, newPassword string) error
    ChangePassword(ctx context.Context, accessToken string, changePasswordData domains.ChangePasswordDTO) error
}
```

Реализован `internal/repositories/auth.Repository`.

---

## TokensRepository

Файл: `internal/interfaces/repositories.go`

```go
type TokensRepository interface {
    Save(ctx context.Context, tokens domains.TokensDTO) error
    Load(ctx context.Context) (*domains.TokensDTO, error)
    Delete(_ context.Context) error
}
```

Реализован `internal/repositories/tokens.Repository`.

---

## UsersRepository

Файл: `internal/interfaces/repositories.go`

```go
type UsersRepository interface {
    GetCurrentUser(ctx context.Context, accessToken string) (*domains.User, error)
    UpdateUser(ctx context.Context, accessToken string, updateUserData domains.UpdateUserDTO) (*domains.User, error)
    SearchUsers(ctx context.Context, filters *domains.UsersFilters, pagination *domains.Pagination) ([]domains.User, error)
}
```

Реализован `internal/repositories/users.Repository`.

---

## ChatsRepository

Файл: `internal/interfaces/repositories.go`

```go
type ChatsRepository interface {
    GetUserChats(ctx context.Context, accessToken string, pagination *domains.Pagination) ([]domains.Chat, error)
    CreateChat(ctx context.Context, accessToken string, chat domains.Chat) (*domains.Chat, error)
}
```

Реализован `internal/repositories/chats.Repository`.

---

## MessagesRepository

Файл: `internal/interfaces/repositories.go`

```go
type MessagesRepository interface {
    GetChatMessages(ctx context.Context, accessToken string, chatID uint64, pagination *domains.Pagination) ([]domains.Message, error)
    GetMessageByID(ctx context.Context, accessToken string, messageID uint64) (*domains.Message, error)
    UpdateMessage(ctx context.Context, accessToken string, dto domains.UpdateMessageDTO) error
    DeleteMessage(ctx context.Context, accessToken string, dto domains.DeleteMessageDTO) error
}
```

Реализован `internal/repositories/messages.Repository`. Реакции вынесены в отдельный `ReactionsRepository`.

---

## ReactionsRepository

Файл: `internal/interfaces/repositories.go`

```go
type ReactionsRepository interface {
    // Публичный эндпоинт /api/reactions — accessToken не нужен.
    ListReactions(ctx context.Context) ([]domains.Reaction, error)
    AddMessageReaction(ctx context.Context, accessToken string, dto domains.MessageReactionDTO) error
    RemoveMessageReaction(ctx context.Context, accessToken string, dto domains.MessageReactionDTO) error
}
```

Реализован `internal/repositories/reactions.Repository`. `ListReactions` бьёт в публичный роут — cookie с access-token не отправляется.

---

## WebSocketsRepository

Файл: `internal/interfaces/repositories.go`

```go
type WebSocketsRepository interface {
    Connect(ctx context.Context, accessToken string) error
    Close() error
    ReadMessage(ctx context.Context) (*domains.Message, error)
    WriteMessage(ctx context.Context, message domains.Message) error
}
```

Реализован `internal/repositories/ws.Repository`.

---

## SettingsRepository

Файл: `internal/interfaces/repositories.go`

```go
type SettingsRepository interface {
    Save(ctx context.Context, settings domains.Settings) error
    Load(ctx context.Context) (*domains.Settings, error)
    Delete(_ context.Context) error
}
```

Реализован `internal/repositories/settings.Repository`.

---

## ErrorsMapper

Файл: `internal/interfaces/errors.go`

```go
type ErrorsMapper interface {
    Map(err error) error
}
```

Реализован `internal/errors.Mapper`. Преобразует технические ошибки в пользовательские строки на русском языке.

---

## HTTPClient

Файл: `internal/interfaces/http.go`

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}
```

Реализован стандартным `*http.Client`. Интерфейс введён для возможности подмены в тестах.

---

## Window

Файл: `internal/interfaces/window.go`

```go
type Window interface {
    Build(content fyne.CanvasObject)
    Show()
    Close()
}
```

Используется только в легаси v1 (Fyne). В активной версии v2 (Wails) не применяется.
