# Пакет internal/interfaces

## Назначение

Определение всех интерфейсов приложения. Используется для инверсии зависимостей между слоями. Моки генерируются через `go:generate mockgen`.

## Интерфейсы

### `Application` (`application.go`)

Жизненный цикл приложения.

| Метод | Описание |
|-------|----------|
| `Startup(ctx context.Context)` | Инициализация при запуске |
| `Shutdown(ctx context.Context)` | Очистка при завершении |
| `BindHandlers() []any` | Возвращает хендлеры для Wails-биндингов |

### `Handler` (`handler.go`)

Базовый интерфейс хендлера.

| Метод | Описание |
|-------|----------|
| `SetContext(ctx context.Context)` | Устанавливает контекст Wails |
| `StartListening()` | Запуск фоновых слушателей (WebSocket и др.) |
| `StopListening()` | Остановка фоновых слушателей |

### `UseCases` (`usecases.go`)

Бизнес-логика приложения. Методы сгруппированы по доменам:

**Auth:**

| Метод | Описание |
|-------|----------|
| `Authenticate(ctx) (*User, error)` | Аутентификация по сохранённым токенам |
| `GetCurrentUser(ctx) (*User, error)` | Получение текущего пользователя |
| `Login(ctx, LoginDTO) (*User, error)` | Вход |
| `Logout(ctx) error` | Выход |
| `Register(ctx, RegisterDTO) (*User, error)` | Регистрация |
| `RefreshTokens(ctx) (*TokensDTO, error)` | Обновление токенов |
| `SendVerifyEmailMessage(ctx, email) error` | Отправка письма подтверждения email |
| `SendForgetPasswordMessage(ctx, email) error` | Отправка письма сброса пароля |
| `ForgetPassword(ctx, token, newPassword) error` | Сброс пароля по токену |
| `ChangePassword(ctx, ChangePasswordDTO) error` | Смена пароля |

**Messaging:**

| Метод | Описание |
|-------|----------|
| `SendMessage(ctx, Message) error` | Отправка сообщения через WebSocket |
| `ReadEvent(ctx) (*WSEvent, error)` | Чтение WS-события из WebSocket (envelope с type + payload) |
| `GetMessageByID(ctx, messageID uint64) (*Message, error)` | Получение сообщения по ID |
| `GetChatMessages(ctx, chatID, *Pagination) ([]Message, error)` | Получение сообщений чата |
| `DeleteMessage(ctx, DeleteMessageDTO) error` | Удаление сообщения (для себя или для всех) |

**Chats:**

| Метод | Описание |
|-------|----------|
| `CreateChat(ctx, Chat) (*Chat, error)` | Создание чата |
| `GetUserChats(ctx, *Pagination) ([]Chat, error)` | Получение чатов пользователя |

**Users:**

| Метод | Описание |
|-------|----------|
| `UpdateUser(ctx, UpdateUserDTO) (*User, error)` | Обновление профиля |
| `SearchUsers(ctx, *UsersFilters, *Pagination) ([]User, error)` | Поиск пользователей |

**Settings:**

| Метод | Описание |
|-------|----------|
| `GetTheme(ctx) ThemeType` | Получение текущей темы |
| `SetTheme(ctx, ThemeType) error` | Установка темы |
| `GetSettings(ctx) (*Settings, error)` | Получение настроек |
| `UpdateSettings(ctx, Settings) (*Settings, error)` | Обновление настроек |

### Репозитории (`repositories.go`)

**`AuthRepository`** — авторизация через HTTP API:

| Метод | Описание |
|-------|----------|
| `Register(ctx, RegisterDTO) (*User, error)` | Регистрация |
| `Login(ctx, LoginDTO) (*TokensDTO, error)` | Вход |
| `Logout(ctx, accessToken) error` | Выход |
| `RefreshTokens(ctx, refreshToken) (*TokensDTO, error)` | Обновление токенов |
| `SendVerifyEmailMessage(ctx, email) error` | Отправка подтверждения email |
| `SendForgetPasswordMessage(ctx, email) error` | Отправка письма сброса пароля |
| `ForgetPassword(ctx, token, newPassword) error` | Сброс пароля |
| `ChangePassword(ctx, accessToken, ChangePasswordDTO) error` | Смена пароля |

**`TokensRepository`** — локальное хранение токенов:

| Метод | Описание |
|-------|----------|
| `Save(ctx, TokensDTO) error` | Сохранение токенов |
| `Load(ctx) (*TokensDTO, error)` | Загрузка токенов |
| `Delete(ctx) error` | Удаление токенов |

**`UsersRepository`** — пользователи через HTTP API:

| Метод | Описание |
|-------|----------|
| `GetCurrentUser(ctx, accessToken) (*User, error)` | Получение текущего пользователя |
| `UpdateUser(ctx, accessToken, UpdateUserDTO) (*User, error)` | Обновление пользователя |
| `SearchUsers(ctx, *UsersFilters, *Pagination) ([]User, error)` | Поиск пользователей |

**`ChatsRepository`** — чаты через HTTP API:

| Метод | Описание |
|-------|----------|
| `GetUserChats(ctx, accessToken, *Pagination) ([]Chat, error)` | Получение чатов |
| `CreateChat(ctx, accessToken, Chat) (*Chat, error)` | Создание чата |

**`MessagesRepository`** — сообщения через HTTP API:

| Метод | Описание |
|-------|----------|
| `GetChatMessages(ctx, accessToken, chatID, *Pagination) ([]Message, error)` | Получение сообщений |
| `GetMessageByID(ctx, accessToken, messageID) (*Message, error)` | Получение сообщения по ID |
| `DeleteMessage(ctx, accessToken, DeleteMessageDTO) error` | Удаление сообщения |

**`WebSocketsRepository`** — WebSocket-соединение:

| Метод | Описание |
|-------|----------|
| `Connect(ctx, accessToken) error` | Подключение |
| `Close() error` | Закрытие соединения |
| `ReadEvent(ctx) (*WSEvent, error)` | Чтение WS-события (envelope с type + payload) |
| `WriteMessage(ctx, Message) error` | Отправка сообщения |

**`SettingsRepository`** — настройки через HTTP API:

| Метод | Описание |
|-------|----------|
| `GetSettings(ctx, accessToken) (*Settings, error)` | Получение настроек |
| `UpdateSettings(ctx, accessToken, Settings) (*Settings, error)` | Обновление настроек |

### `ErrorsMapper` (`errors.go`)

| Метод | Описание |
|-------|----------|
| `Map(err error) error` | Маппинг технической ошибки в пользовательскую |

### `HTTPClient` (`http.go`)

| Метод | Описание |
|-------|----------|
| `Do(req *http.Request) (*http.Response, error)` | Выполнение HTTP-запроса |

### `Window` (`window.go`)

| Метод | Описание |
|-------|----------|
| `Build(content fyne.CanvasObject)` | Построение окна |
| `Show()` | Отображение окна |
| `Close()` | Закрытие окна |

## Зависимости

- `context`, `net/http`
- `fyne.io/fyne/v2`
- `github.com/DKhorkov/kfcGUI/internal/domains`
