# Обработка ошибок

Пакет: `internal/errors/`

## Принцип

Внутренние ошибки (технические, на английском языке) определены в отдельных файлах по доменам. Маппер (`Mapper`) преобразует их в пользовательские строки на русском языке. Хендлеры получают от UseCase уже смапированную ошибку и возвращают её напрямую во фронтенд.

Цепочка: `repository error` -> `usecases.errorsMapper.Map(err)` -> `handler` -> `frontend`.

## Ошибки по доменам

### Аутентификация (`internal/errors/auth.go`)

| Ошибка | Сообщение |
|---|---|
| `ErrRegister` | `registration failed` |
| `ErrLogin` | `login failed` |
| `ErrLogout` | `logout failed` |
| `ErrRefreshTokens` | `refresh tokens failed` |
| `ErrInvalidLogin` | `invalid login` |
| `ErrInvalidPassword` | `invalid password` |
| `ErrInvalidUsername` | `invalid username` |
| `ErrInvalidEmail` | `invalid email` |
| `ErrPasswordDoesNotMatch` | `passwords does not match` |
| `ErrInvalidForgetPasswordToken` | `invalid jwt token: invalid forget_password_token` |
| `ErrTokenExpired` | `token expired` |
| `ErrEmailAlreadyConfirmed` | `email already confirmed` |
| `ErrEmailNotConfirmed` | `email not confirmed` |
| `ErrEmailAlreadyExists` | `email already exist` |
| `ErrWrongPassword` | `wrong password` |
| `ErrAccessTokenDoesNotBelongToRefreshToken` | `access token does not belong to refresh token` |
| `ErrInvalidJwtToken` | `invalid jwt token` |
| `ErrValidationFailed` | `validation failed` |
| `ErrNewPasswordEqualToOldPassword` | `new password equal to old password: ...` |

### Чаты (`internal/errors/chats.go`)

| Ошибка | Сообщение |
|---|---|
| `ErrCreateChat` | `failed to create chat` |
| `ErrGetUserChats` | `failed to get user chats` |
| `ErrGetChatMessages` | `failed to get chat messages` |
| `ErrInvalidChat` | `invalid chat` |
| `ErrUserIsNotChatMember` | `user is not a chat member` |
| `ErrChatNotFound` | `chat not found` |
| `ErrChatAlreadyExists` | `chat already exists` |

### Пользователи (`internal/errors/users.go`)

| Ошибка | Сообщение |
|---|---|
| `ErrUserNotFound` | `user not found` |
| `ErrUserAlreadyExists` | `user already exists` |
| `ErrUpdateUsername` | (сообщение об ошибке уникальности из PostgreSQL) |

### WebSocket (`internal/errors/ws.go`)

| Ошибка | Сообщение |
|---|---|
| `ErrWebsocket` | `websocket error` |
| `ErrWebsocketClosed` | `websocket closed` |

### Лимиты (`internal/errors/limit.go`)

| Ошибка | Сообщение |
|---|---|
| `ErrLimitExceeded` | `limit exceeded` |

## Маппер (`internal/errors/mapper.go`)

```go
type Mapper struct {
    mapping []mappingEntry
}

func New() *Mapper
func (m *Mapper) Map(err error) error
```

Реализует интерфейс `interfaces.ErrorsMapper`.

### Алгоритм маппинга

1. Если `err == nil` — возвращает `nil`.
2. Получает строку ошибки через `err.Error()`.
3. Перебирает таблицу маппинга в порядке убывания длины ключа (более специфичные ключи проверяются раньше).
4. Если строка ошибки **содержит** ключ (`strings.Contains`), возвращает соответствующую пользовательскую ошибку.
5. Если ни один ключ не совпал, возвращает `ErrDefault` — `"Что-то пошло не так..."`.

Сортировка по убывающей длине ключа важна, так как некоторые ошибочные строки могут быть подстроками других (например, `invalid jwt token` является подстрокой `invalid jwt token: invalid forget_password_token`).

### Пользовательские сообщения (русский язык)

| Внутренняя ошибка | Пользовательское сообщение |
|---|---|
| `ErrUserNotFound` | Такого пользователя не существует |
| `ErrUserAlreadyExists` | Пользователь с такой почтой или логином уже существует |
| `ErrUpdateUsername` | Пользователь с таким логином уже существует |
| `ErrEmailAlreadyExists` | Этот почтовый адрес уже занят |
| `ErrLogin` / `ErrWrongPassword` | Неверный логин или пароль |
| `ErrEmailNotConfirmed` | Почта не была подтверждена |
| `ErrEmailAlreadyConfirmed` | Эта почта уже подтверждена |
| `ErrInvalidJwtToken` / `ErrValidationFailed` | Ошибка авторизации |
| `ErrInvalidPassword` | Пароль должен быть на латинице, не менее 8 символов... |
| `ErrInvalidLogin` | Некорректный email или логин... |
| `ErrInvalidUsername` | Логин должен быть не менее 5 символов... |
| `ErrInvalidEmail` | Некорректный адрес электронной почты |
| `ErrPasswordDoesNotMatch` | Пароли не совпадают |
| `ErrInvalidForgetPasswordToken` | Некорректный код для сброса пароля |
| `ErrNewPasswordEqualToOldPassword` | Старый пароль не может быть использован в качестве нового пароля |
| `ErrTokenExpired` | Токен устарел |
| `ErrInvalidChat` | Неверный чат |
| `ErrUserIsNotChatMember` | У вас нет доступа к этому чату |
| `ErrChatNotFound` | Чат не найден |
| `ErrChatAlreadyExists` | Такой чат уже существует |
| `ErrGetUserChats` | Не удалось получить чаты пользователя |
| `ErrCreateChat` | Не удалось создать чат |
| `ErrGetChatMessages` | Не удалось получить сообщения для чата |
| `ErrWebsocket` / `ErrWebsocketClosed` | Что-то пошло не так... |
| `ErrLimitExceeded` | Превышен лимит. Попробуйте позже |
