# Пакет internal/domains

## Назначение

Доменные модели и DTO приложения. Описывает сущности пользователей, чатов, сообщений, авторизации, настроек и пагинации.

## Типы

### Пользователи (`users.go`)

| Тип | Описание |
|-----|----------|
| `User` | Пользователь: `ID`, `Username`, `Email`, `EmailConfirmed`, `Password`, `CreatedAt`, `UpdatedAt`, `AvatarPath *string` (URL аватара, nullable) |
| `UpdateUserDTO` | DTO обновления пользователя: `Username *string` |
| `UsersFilters` | Фильтры поиска пользователей: `Username *string` |

### Авторизация (`auth.go`)

| Тип | Описание |
|-----|----------|
| `LoginDTO` | Данные входа: `Login`, `Password` |
| `RegisterDTO` | Данные регистрации: `Username`, `Email`, `Password` |
| `TokensDTO` | Пара токенов: `AccessToken`, `RefreshToken` |
| `ForgetPasswordDTO` | Сброс пароля: `NewPassword` |
| `ChangePasswordDTO` | Смена пароля: `NewPassword`, `OldPassword` |
| `SendVerifyEmailMessageDTO` | Отправка письма подтверждения: `Email` |
| `SendForgetPasswordMessageDTO` | Отправка письма сброса пароля: `Email` |

### Чаты (`chat.go`)

| Тип | Описание |
|-----|----------|
| `ChatType` | Тип чата (`string`). Значения: `ChatTypePrivate` (`"private"`), `ChatTypeGroup` (`"group"`) |
| `Chat` | Чат: `ID`, `Title`, `Description`, `Type`, `CreatedAt`, `UpdatedAt`, `UnreadCount uint64` (число непрочитанных сообщений; для определения «есть непрочитанные» — `UnreadCount > 0`), `Members`, `Messages` |
| `CreateChatDTO` | DTO создания чата: `Title`, `Description`, `Type`, `MemberIDs` |

**Переменные:**

- `ChatTypes` — допустимые типы чатов (`[]ChatType`)

**Методы:**

| Метод | Описание |
|-------|----------|
| `(*Chat).IsValid() bool` | Проверяет валидность чата: корректный тип, достаточное количество участников, для приватного — ровно 2 участника |
| `(*CreateChatDTO).IsValid() bool` | Проверяет валидность DTO создания: корректный тип, для приватного — ровно 1 `MemberID` (текущий пользователь добавляется на бэкенде) |

### Сообщения (`message.go`)

| Тип | Описание |
|-----|----------|
| `Message` | Сообщение: `ID`, `ChatID`, `Sender` (`User`), `Text`, `CreatedAt`, `UpdatedAt`, `IsRead`, `ReplyToMessage *Message` |
| `DeleteMessageDTO` | DTO удаления сообщения: `MessageID`, `ForAll` (`UserID` исключён из JSON через `json:"-"`) |

**Функции и методы:**

| Функция/метод | Описание |
|---------------|----------|
| `NewMessage() *Message` | Создаёт пустое сообщение (builder-паттерн) |
| `(*Message).From(user User) *Message` | Устанавливает отправителя |
| `(*Message).Received() *Message` | Устанавливает `CreatedAt` в текущее время |
| `(*Message).Updated() *Message` | Устанавливает `UpdatedAt` в текущее время |

### WebSocket-события (`ws_event.go`)

| Тип | Описание |
|-----|----------|
| `WSEventType` | Тип WS-события (`string`). Значения: `WSEventNewMessage` (`"new_message"`), `WSEventMessageDeleted` (`"message_deleted"`), `WSEventMessageEdited` (`"message_edited"`) |
| `WSEvent` | WS-событие: `Type WSEventType`, `Payload json.RawMessage` |
| `MessageDeletedPayload` | Полезная нагрузка удаления: `MessageID uint64`, `ChatID uint64` |
| `MessageEditedPayload` | Полезная нагрузка редактирования: `MessageID uint64`, `ChatID uint64` |

### Пагинация (`pagination.go`)

| Тип | Описание |
|-----|----------|
| `Pagination` | Параметры пагинации: `Limit *uint64`, `Offset *uint64` |

### Настройки (`setings.go`)

| Тип | Описание |
|-----|----------|
| `ThemeType` | Тип темы (`int`). Значения: `ThemeLight` (0), `ThemeDark` (1) |
| `NotificationConsent` | Битовая маска согласий на уведомления (`int`). Флаги: `ConsentNewMessage` (1) |
| `Settings` | Настройки пользователя: `Theme`, `EmailConsents`, `WebPushConsents` |

## Зависимости

- `time`, `slices`
