# Бизнес-логика (UseCases)

Файл: `internal/usecases/usecases.go`

## Структура

```go
type UseCases struct {
    users        interfaces.UsersRepository
    chats        interfaces.ChatsRepository
    auth         interfaces.AuthRepository
    tokens       interfaces.TokensRepository
    settings     interfaces.SettingsRepository
    ws           interfaces.WebSocketsRepository
    logger       logging.Logger
    errorsMapper interfaces.ErrorsMapper
}
```

Реализует интерфейс `interfaces.UseCases`.

## Конструктор

```go
func New(
    users interfaces.UsersRepository,
    chats interfaces.ChatsRepository,
    auth interfaces.AuthRepository,
    tokens interfaces.TokensRepository,
    settings interfaces.SettingsRepository,
    ws interfaces.WebSocketsRepository,
    logger logging.Logger,
    errorsMapper interfaces.ErrorsMapper,
) *UseCases
```

## Общий принцип

Все методы:
1. Выполняют необходимые действия через репозитории.
2. При ошибке логируют её через `logging.LogErrorContext`.
3. Оборачивают ошибку через `u.errorsMapper.Map(err)` перед возвратом вызывающему.

## Методы аутентификации

### `Authenticate(ctx) (*User, error)`
Обновляет токены, затем получает текущего пользователя. Используется для восстановления сессии при старте приложения.

### `GetCurrentUser(ctx) (*User, error)`
Аналогично `Authenticate` — обновляет токены и возвращает пользователя.

### `RefreshTokens(ctx) (*TokensDTO, error)`
- Загружает токены из локального хранилища (`tokens.Load`).
- Запрашивает у сервера новую пару токенов (`auth.RefreshTokens`).
- Сохраняет новые токены локально (`tokens.Save`).

Используется внутренне и хендлером чата (периодическое обновление).

### `Login(ctx, LoginDTO) (*User, error)`
- Выполняет вход через `auth.Login`, получает пару токенов.
- Сохраняет токены локально.
- Получает и возвращает текущего пользователя.

### `Logout(ctx) error`
- Загружает токены.
- Выполняет выход через `auth.Logout`.
- Удаляет токены из локального хранилища.
- Закрывает WebSocket-соединение.

### `Register(ctx, RegisterDTO) (*User, error)`
Делегирует регистрацию в `auth.Register`. Токены не сохраняются — пользователь должен войти после регистрации.

### `SendVerifyEmailMessage(ctx, email) error`
Делегирует в `auth.SendVerifyEmailMessage`.

### `SendForgetPasswordMessage(ctx, email) error`
Делегирует в `auth.SendForgetPasswordMessage`.

### `ForgetPassword(ctx, token, newPassword) error`
Делегирует в `auth.ForgetPassword`.

### `ChangePassword(ctx, ChangePasswordDTO) error`
- Загружает токены.
- Делегирует в `auth.ChangePassword` с передачей `accessToken`.

## Методы сообщений

### `SendMessage(ctx, Message) error`
- Загружает токены.
- Устанавливает WebSocket-соединение (`ws.Connect`), если не установлено.
- Отправляет сообщение через `ws.WriteMessage`.

### `ReadMessage(ctx) (*Message, error)`
- Загружает токены.
- Устанавливает WebSocket-соединение.
- Читает следующее сообщение из буфера (`ws.ReadMessage`). Блокируется до получения.

### `GetChatMessages(ctx, chatID, pagination) ([]Message, error)`
- Загружает токены.
- Получает сообщения через REST (`chats.GetChatMessages`).
- Приводит время каждого сообщения к локальному часовому поясу пользователя (`common.Timezone`).

## Методы чатов

### `CreateChat(ctx, Chat) (*Chat, error)`
- Загружает токены.
- Создаёт чат через REST (`chats.CreateChat`).

### `GetUserChats(ctx, pagination) ([]Chat, error)`
- Загружает токены.
- Получает список чатов через REST.

## Методы пользователей

### `UpdateUser(ctx, UpdateUserDTO) (*User, error)`
- Загружает токены.
- Обновляет пользователя через REST.

### `SearchUsers(ctx, filters, pagination) ([]User, error)`
- Загружает токены.
- Получает текущего пользователя.
- Выполняет поиск через REST.
- Фильтрует из результатов текущего пользователя (чтобы пользователь не видел сам себя в поиске).

## Методы настроек

### `GetTheme(ctx) ThemeType`
Загружает настройки из файла. При ошибке (файл не существует) возвращает `ThemeLight`.

### `SetTheme(ctx, ThemeType) error`
- Пытается загрузить существующие настройки.
- Устанавливает тему.
- Сохраняет настройки в файл.
