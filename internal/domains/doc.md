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
| `Message` | Сообщение: `ID`, `ChatID`, `Sender` (`User`), `Text`, `CreatedAt`, `UpdatedAt`, `IsRead`, `ReplyToMessage *Message`, `Reactions []MessageReactionSummary` |
| `DeleteMessageDTO` | DTO удаления сообщения: `MessageID`, `ForAll` (`UserID` исключён из JSON через `json:"-"`) |

### Реакции (`reaction.go`)

| Тип | Описание |
|-----|----------|
| `Reaction` | Элемент справочника: `ID`, `Emoji` |
| `MessageReactionSummary` | Агрегат на сообщении: `Reaction`, `UserIDs []uint64` — кто поставил. `count` вычисляется на фронте как `len(UserIDs)` |
| `MessageReactionDTO` | DTO API: `MessageID` (из URL, `json:"-"`), `ReactionID` (в body) |

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
| `WSEventType` | Тип WS-события (`string`). Значения: `WSEventNewMessage` (`"new_message"`), `WSEventMessageDeleted` (`"message_deleted"`), `WSEventMessageEdited` (`"message_edited"`), `WSEventReactionAdded` (`"reaction_added"`), `WSEventReactionRemoved` (`"reaction_removed"`) |
| `WSEvent` | WS-событие: `Type WSEventType`, `Payload json.RawMessage` |
| `MessageDeletedPayload` | Полезная нагрузка удаления: `MessageID uint64`, `ChatID uint64` |
| `MessageEditedPayload` | Полезная нагрузка редактирования: `MessageID uint64`, `ChatID uint64` |
| `ReactionAddedPayload` | Постановка реакции: `MessageID`, `ChatID`, `UserID`, `ReactionID`, `Emoji` |
| `ReactionRemovedPayload` | Снятие реакции: `MessageID`, `ChatID`, `UserID`, `ReactionID` |

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
