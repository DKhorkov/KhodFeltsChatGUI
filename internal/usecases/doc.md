# Пакет `internal/usecases`

## Назначение

Слой бизнес-логики приложения. Реализует интерфейс `interfaces.UseCases`, оркестрируя вызовы к репозиториям (auth, users, chats, tokens, settings, websockets). Все методы логируют ошибки и маппят их через `ErrorsMapper` в пользовательские сообщения.

## Типы

| Тип | Описание |
|-----|----------|
| `UseCases` | Основная структура, хранит ссылки на репозитории, логгер и маппер ошибок |

## Конструктор

| Функция | Сигнатура | Описание |
|---------|-----------|----------|
| `New` | `New(users, chats, messages, auth, tokens, settings, ws, logger, errorsMapper) *UseCases` | Создает экземпляр `UseCases` со всеми зависимостями |

## Методы

### Аутентификация и авторизация

| Метод | Сигнатура | Описание |
|-------|-----------|----------|
| `Authenticate` | `(ctx) (*domains.User, error)` | Обновляет токены и возвращает текущего пользователя |
| `GetCurrentUser` | `(ctx) (*domains.User, error)` | Обновляет токены и возвращает текущего пользователя (аналогичен `Authenticate`) |
| `RefreshTokens` | `(ctx) (*domains.TokensDTO, error)` | Загружает токены из хранилища, обновляет через API и сохраняет обратно |
| `Login` | `(ctx, domains.LoginDTO) (*domains.User, error)` | Выполняет вход, сохраняет токены, возвращает пользователя |
| `Logout` | `(ctx) error` | Выполняет выход: вызывает API, удаляет токены, закрывает WebSocket-соединение |
| `Register` | `(ctx, domains.RegisterDTO) (*domains.User, error)` | Регистрирует нового пользователя |

### Восстановление пароля

| Метод | Сигнатура | Описание |
|-------|-----------|----------|
| `SendVerifyEmailMessage` | `(ctx, email string) error` | Отправляет письмо для подтверждения email |
| `SendForgetPasswordMessage` | `(ctx, email string) error` | Отправляет письмо для сброса пароля |
| `ChangePassword` | `(ctx, domains.ChangePasswordDTO) error` | Меняет пароль авторизованного пользователя |
| `ForgetPassword` | `(ctx, forgetPasswordToken, newPassword string) error` | Устанавливает новый пароль по токену сброса |

### Чаты и сообщения

| Метод | Сигнатура | Описание |
|-------|-----------|----------|
| `CreateChat` | `(ctx, domains.Chat) (*domains.Chat, error)` | Создает новый чат |
| `GetUserChats` | `(ctx, *domains.Pagination) ([]domains.Chat, error)` | Получает чаты текущего пользователя с пагинацией |
| `GetMessageByID` | `(ctx, messageID uint64) (*domains.Message, error)` | Получает сообщение по ID; конвертирует время в локальную таймзону |
| `GetChatMessages` | `(ctx, chatID uint64, *domains.Pagination) ([]domains.Message, error)` | Получает сообщения чата; конвертирует время в локальную таймзону |
| `SendMessage` | `(ctx, domains.Message) error` | Подключается к WebSocket (если не подключен) и отправляет сообщение |
| `ReadEvent` | `(ctx) (*domains.WSEvent, error)` | Подключается к WebSocket и читает WS-событие (envelope) |
| `DeleteMessage` | `(ctx, domains.DeleteMessageDTO) error` | Удаление сообщения через HTTP API |
| `UpdateMessage` | `(ctx, domains.UpdateMessageDTO) error` | Редактирование сообщения: загружает токены, вызывает `messages.UpdateMessage`, маппит ошибки |

### Пользователи

| Метод | Сигнатура | Описание |
|-------|-----------|----------|
| `UpdateUser` | `(ctx, domains.UpdateUserDTO) (*domains.User, error)` | Обновляет данные пользователя |
| `SearchUsers` | `(ctx, *domains.UsersFilters, *domains.Pagination) ([]domains.User, error)` | Ищет пользователей, исключая текущего из результатов |

### Настройки

| Метод | Сигнатура | Описание |
|-------|-----------|----------|
| `GetTheme` | `(ctx) domains.ThemeType` | Возвращает текущую тему; при ошибке возвращает `ThemeLight` по умолчанию |
| `SetTheme` | `(ctx, domains.ThemeType) error` | Загружает текущие настройки, обновляет тему и сохраняет |
| `GetSettings` | `(ctx) (*domains.Settings, error)` | Возвращает настройки пользователя |
| `UpdateSettings` | `(ctx, domains.Settings) (*domains.Settings, error)` | Обновляет настройки пользователя |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/common` -- таймзона для конвертации времени сообщений
- `github.com/DKhorkov/kfcGUI/internal/domains` -- доменные модели и DTO
- `github.com/DKhorkov/kfcGUI/internal/interfaces` -- интерфейсы репозиториев, `UseCases`, `ErrorsMapper`
- `github.com/DKhorkov/libs/logging` -- логирование
