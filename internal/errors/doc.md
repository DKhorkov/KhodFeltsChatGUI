# Пакет internal/errors

## Назначение

Определение ошибок приложения и их маппинг в пользовательские сообщения на русском языке. Ошибки сгруппированы по доменам: авторизация, пользователи, чаты, настройки, WebSocket, лимиты.

## Ошибки

### Авторизация (`auth.go`)

| Переменная | Сообщение |
|------------|-----------|
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
| `ErrNewPasswordEqualToOldPassword` | `new password equal to old password` |

### Пользователи (`users.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrUserNotFound` | `user not found` |
| `ErrUserAlreadyExists` | `user already exists` |
| `ErrUpdateUsername` | Ошибка уникального ограничения PostgreSQL на `users_username_key` |

### Чаты (`chats.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrCreateChat` | `failed to create chat` |
| `ErrGetUserChats` | `failed to get user chats` |
| `ErrGetChatMessages` | `failed to get chat messages` |
| `ErrInvalidChat` | `invalid chat` |
| `ErrUserIsNotChatMember` | `user is not a chat member` |
| `ErrChatNotFound` | `chat not found` |
| `ErrChatAlreadyExists` | `chat already exists` |

### Настройки (`settings.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrSettingsNotFound` | `settings not found` |
| `ErrWebPushSubscriptionNotFound` | `web-push subscription not found` |
| `ErrWebPushSubscriptionExpired` | `web-push subscription expired` |

### Сообщения (`messages.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrMessageNotFound` | `message not found` |
| `ErrNotMessageAuthor` | `only message author can perform this action` |

### Реакции (`reactions.go`)

Тексты sentinel'ов совпадают с серверными (KFC `internal/errors/reactions.go`). Репозиторий `messages` подменяет сырое тело ответа на sentinel по HTTP-коду (409 → `ErrReactionAlreadyExists`, 404 → `ErrReactionNotFound`), после чего маппер переводит их в русские сообщения.

| Переменная | Сообщение |
|------------|-----------|
| `ErrReactionAlreadyExists` | `reaction already exists on this message for this user` |
| `ErrReactionNotFound` | `reaction not found in dictionary` |
| `ErrReactionNotSet` | `reaction was not set on this message for this user` |

### Файловое хранилище (`file_storage.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrFileNotFound` | `file not found` |
| `ErrInvalidImageFormat` | `invalid image format: supported formats are JPEG, PNG, WebP, GIF` |
| `ErrFileTooLarge` | `file too large` |

### WebSocket (`ws.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrWebsocket` | `websocket error` |
| `ErrWebsocketClosed` | `websocket closed` |

### Лимиты (`limit.go`)

| Переменная | Сообщение |
|------------|-----------|
| `ErrLimitExceeded` | `limit exceeded` |

## Маппер ошибок (`mapper.go`)

### `Mapper`

Маппер технических ошибок (на английском) в пользовательские сообщения (на русском). Ключи сортируются по убыванию длины для приоритета более специфичных совпадений.

| Функция/метод | Описание |
|---------------|----------|
| `New() *Mapper` | Создаёт маппер с предзаполненной таблицей соответствий |
| `(*Mapper).Map(err error) error` | Преобразует ошибку в пользовательское сообщение. Ищет подстроку ошибки в таблице маппинга. Возвращает `ErrDefault` при отсутствии совпадения |

**`ErrDefault`** — дефолтная ошибка: `"Что-то пошло не так..."`

### Русские сообщения реакций (в маппере)

| Sentinel | Русский текст |
|----------|---------------|
| `ErrReactionAlreadyExists` | `Реакция уже поставлена` |
| `ErrReactionNotFound` | `Такой реакции не существует` |
| `ErrReactionNotSet` | `Реакция не была поставлена` |

## Зависимости

- `errors`, `sort`, `strings`
