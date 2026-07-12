# Репозитории

Пакет: `internal/repositories/`

## base.Repository

Файл: `internal/repositories/base/repository.go`

Встраиваемая структура, предоставляющая общий вспомогательный метод для HTTP-репозиториев.

```go
type Repository struct {
    logger logging.Logger
}

func (r *Repository) CloseBody(ctx context.Context, body io.ReadCloser)
```

`CloseBody` безопасно закрывает тело HTTP-ответа, логируя ошибку при её возникновении. Используется в defer-блоках всех HTTP-репозиториев.

Примечание: в `auth`, `users` и `chats` репозиториях `base.Repository` встраивается по значению (не через конструктор), поэтому поле `logger` остаётся нулевым — логирование ошибок закрытия тела в этих репозиториях не производится.

---

## auth.Repository

Файл: `internal/repositories/auth/repository.go`

HTTP-репозиторий для операций аутентификации. Взаимодействует с REST API сервера.

### Конструктор

```go
func New(httpClient interfaces.HTTPClient, baseURL string) *Repository
```

### Методы и эндпоинты

| Метод | HTTP-метод | Эндпоинт | Описание |
|---|---|---|---|
| `Register(ctx, RegisterDTO)` | POST | `/users` | Регистрирует пользователя, возвращает `*User` |
| `Login(ctx, LoginDTO)` | POST | `/sessions` | Логинит пользователя, возвращает токены из cookie |
| `Logout(ctx, accessToken)` | DELETE | `/sessions` | Завершает сессию, accessToken передаётся в cookie |
| `RefreshTokens(ctx, refreshToken)` | PUT | `/sessions` | Обновляет пару токенов, refreshToken в cookie |
| `SendVerifyEmailMessage(ctx, email)` | POST | `/users/email/verify` | Отправляет письмо для верификации email |
| `SendForgetPasswordMessage(ctx, email)` | POST | `/users/password/forget` | Отправляет письмо для сброса пароля |
| `ForgetPassword(ctx, token, newPassword)` | POST | `/users/password/forget/{token}` | Применяет новый пароль |
| `ChangePassword(ctx, accessToken, ChangePasswordDTO)` | POST | `/users/password/change` | Меняет пароль, accessToken в cookie |

Токены передаются и извлекаются через HTTP-куки (`accessToken`, `refreshToken`). При входе и обновлении токенов они читаются из `resp.Cookies()`. Если куки отсутствуют — возвращается соответствующая ошибка.

Все методы используют `sync.Mutex` для потокобезопасности.

---

## users.Repository

Файл: `internal/repositories/users/repository.go`

HTTP-репозиторий для операций с пользователями.

### Конструктор

```go
func New(httpClient interfaces.HTTPClient, baseURL string) *Repository
```

### Методы и эндпоинты

| Метод | HTTP-метод | Эндпоинт | Описание |
|---|---|---|---|
| `GetCurrentUser(ctx, accessToken)` | GET | `/users/me` | Возвращает текущего пользователя |
| `UpdateUser(ctx, accessToken, UpdateUserDTO)` | PUT | `/users/me` | Обновляет профиль пользователя |
| `SearchUsers(ctx, filters, pagination)` | GET | `/users?username=...&limit=...&offset=...` | Поиск пользователей по фильтрам |

`GetCurrentUser` и `SearchUsers` используют `sync.RWMutex` (читающая блокировка). `UpdateUser` использует полную блокировку.

---

## chats.Repository

Файл: `internal/repositories/chats/repository.go`

HTTP-репозиторий для операций с чатами.

### Конструктор

```go
func New(httpClient interfaces.HTTPClient, baseURL string) *Repository
```

### Методы и эндпоинты

| Метод | HTTP-метод | Эндпоинт | Описание |
|---|---|---|---|
| `GetUserChats(ctx, accessToken, pagination)` | GET | `/chats?limit=...&offset=...` | Список чатов пользователя |
| `CreateChat(ctx, accessToken, Chat)` | POST | `/chats` | Создаёт чат |

Пагинация (`limit`, `offset`) передаётся как query-параметры при наличии. `accessToken` передаётся в cookie.

---

## messages.Repository

Файл: `internal/repositories/messages/repository.go`

HTTP-репозиторий для операций с сообщениями. Реакции вынесены в отдельный `reactions.Repository`.

### Конструктор

```go
func New(httpClient interfaces.HTTPClient, baseURL string) *Repository
```

### Методы и эндпоинты

| Метод | HTTP-метод | Эндпоинт | Описание |
|---|---|---|---|
| `GetChatMessages(ctx, accessToken, chatID, pagination)` | GET | `/chats/{id}/messages?limit=...&offset=...` | Сообщения чата с пагинацией |
| `GetMessageByID(ctx, accessToken, messageID)` | GET | `/messages/{id}` | Одно сообщение по ID |
| `UpdateMessage(ctx, accessToken, UpdateMessageDTO)` | PUT | `/messages/{id}` | Редактирование текста сообщения |
| `DeleteMessage(ctx, accessToken, DeleteMessageDTO)` | DELETE | `/messages/{id}` | Удаление сообщения (body: `{"forAll": bool}`) |

Read-методы используют `sync.RWMutex.RLock()`, write-методы — `Lock()`.

---

## reactions.Repository

Файл: `internal/repositories/reactions/repository.go`

HTTP-репозиторий для реакций на сообщения. Выделен из `messages.Repository` — реакции это отдельная сущность (справочник + M2M).

### Конструктор

```go
func New(httpClient interfaces.HTTPClient, baseURL string) *Repository
```

### Методы и эндпоинты

| Метод | HTTP-метод | Эндпоинт | Описание |
|---|---|---|---|
| `ListReactions(ctx)` | GET | `/reactions` | Публичный справочник emoji. Cookie `accessToken` **не** отправляется — роут исключён из auth middleware на сервере |
| `AddMessageReaction(ctx, accessToken, MessageReactionDTO)` | POST | `/messages/{id}/reactions` | Body `{"reactionId": N}`. 409 → `ErrReactionAlreadyExists`, 404 → `ErrReactionNotFound` |
| `RemoveMessageReaction(ctx, accessToken, MessageReactionDTO)` | DELETE | `/messages/{id}/reactions/{reactionId}` | Идемпотентно: 204 всегда при успехе |

### Обработка ошибок

`AddMessageReaction` маппит HTTP-статусы в доменные sentinel-ошибки:

- `409 Conflict` → `customerrors.ErrReactionAlreadyExists` — юзер уже поставил эту реакцию.
- `404 Not Found` → `customerrors.ErrReactionNotFound` — реакции с таким ID нет в справочнике.

Остальные не-2xx возвращаются как `errors.New(<body>)`.

---

## tokens.Repository

Файл: `internal/repositories/tokens/repository.go`

Локальное хранилище JWT-токенов. Токены сохраняются в JSON-файл на диске.

### Конструктор

```go
func New() *Repository
```

### Файл хранения

```
{AppDataDir}/tokens.json
```

Права доступа к файлу: `0o600` (только владелец).

### Методы

| Метод | Описание |
|---|---|
| `Save(ctx, TokensDTO) error` | Сериализует токены в JSON и записывает в файл |
| `Load(ctx) (*TokensDTO, error)` | Читает файл и десериализует токены |
| `Delete(ctx) error` | Удаляет файл токенов |

Все методы защищены `sync.RWMutex`.

---

## settings.Repository

Файл: `internal/repositories/settings/repository.go`

Локальное хранилище пользовательских настроек. Аналогично `tokens.Repository`.

### Конструктор

```go
func New() *Repository
```

### Файл хранения

```
{AppDataDir}/settings.json
```

Права доступа: `0o600`.

### Методы

| Метод | Описание |
|---|---|
| `Save(ctx, Settings) error` | Сериализует настройки в JSON и записывает в файл |
| `Load(ctx) (*Settings, error)` | Читает и десериализует настройки |
| `Delete(ctx) error` | Удаляет файл настроек |

---

## ws.Repository

Файл: `internal/repositories/ws/repository.go`

WebSocket-репозиторий для получения и отправки сообщений в реальном времени. Использует библиотеку `github.com/gorilla/websocket`.

### Конструктор

```go
func New(baseURL string, logger logging.Logger) *Repository
```

### Структура

```go
type Repository struct {
    baseURL      string
    logger       logging.Logger
    ws           *websocket.Conn
    mu           sync.Mutex
    messagesChan chan *domains.Message  // буфер 100 сообщений
    errChan      chan error             // буфер 1 ошибки
}
```

### Методы

#### `Connect(ctx, accessToken) error`
Устанавливает WebSocket-соединение с эндпоинтом `/ws`. `accessToken` передаётся в заголовке `Cookie`. Если соединение уже установлено (`r.ws != nil`), метод возвращает `nil` без переподключения. После успешного подключения запускает горутину `readLoop`.

#### `Close() error`
Закрывает WebSocket-соединение и обнуляет `r.ws`. Безопасен при повторном вызове.

#### `ReadMessage(ctx) (*Message, error)`
Неблокирующий (с учётом `ctx.Done()`) прием из `messagesChan`. Возвращает ошибку если:
- контекст отменён;
- в `errChan` появилась ошибка чтения;
- канал `messagesChan` закрыт.

#### `WriteMessage(ctx, Message) error`
Отправляет JSON-сообщение в WebSocket. Устанавливает write deadline в 2 секунды. Корректно обрабатывает закрытие соединения (`CloseNormalClosure`, `CloseGoingAway`, `CloseAbnormalClosure`, `net.ErrClosed`).

### Горутина `readLoop`
Запускается при подключении. Непрерывно читает JSON-сообщения из сокета и помещает их в `messagesChan`. При ошибке чтения (в т.ч. закрытие соединения) кладёт ошибку в `errChan`, закрывает оба канала и вызывает `Close()`.

### Константы

```go
readMessagesBufferSize = 100     // размер буфера входящих сообщений
readErrorsBufferSize   = 1       // размер буфера ошибок
writeDeadline          = 2 * time.Second
```
