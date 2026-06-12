# Доменные модели

Пакет: `internal/domains/`

## Пользователь

Файл: `internal/domains/users.go`

```go
type User struct {
    ID             uint64    `json:"id"`
    Username       string    `json:"username"`
    Email          string    `json:"email"`
    EmailConfirmed bool      `json:"emailConfirmed"`
    Password       string    `json:"password"`
    CreatedAt      time.Time `json:"createdAt"`
    UpdatedAt      time.Time `json:"updatedAt"`
}
```

### DTOs пользователя

```go
type UpdateUserDTO struct {
    Username *string `json:"username,omitempty"`
}

type UsersFilters struct {
    Username *string `json:"username,omitempty"`
}
```

`UsersFilters` используется при поиске пользователей. Все поля опциональны.

## Чат

Файл: `internal/domains/chat.go`

```go
type ChatType string

const (
    ChatTypePrivate ChatType = "private"
    ChatTypeGroup   ChatType = "group"
)

type Chat struct {
    ID          uint64    `json:"id"`
    Title       *string   `json:"title,omitempty"`
    Description *string   `json:"description,omitempty"`
    Type        ChatType  `json:"type"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    UnreadCount uint64    `json:"unreadCount"`
    Members     []User    `json:"members,omitempty"`
    Messages    []Message `json:"messages,omitempty"`
}
```

Метод `Chat.IsValid()` проверяет:
- тип чата присутствует в списке допустимых (`ChatTypePrivate`, `ChatTypeGroup`);
- количество участников >= 1;
- для приватного чата количество участников строго равно 2.

### DTO создания чата

```go
type CreateChatDTO struct {
    Title       *string  `json:"title,omitempty"`
    Description *string  `json:"description,omitempty"`
    Type        ChatType `json:"type"`
    MemberIDs   []uint64 `json:"memberIDs,omitempty"`
}
```

Метод `CreateChatDTO.IsValid()` проверяет:
- корректный тип чата;
- для приватного чата в `MemberIDs` ровно один ID (бэкенд автоматически добавляет текущего пользователя).

## Сообщение

Файл: `internal/domains/message.go`

```go
type Message struct {
    ID        uint64    `json:"id"`
    ChatID    uint64    `json:"chatId"`
    Sender    User      `json:"sender"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    IsRead    bool      `json:"isRead"`
}
```

Конструктор и builder-методы:

```go
func NewMessage() *Message
func (m *Message) From(user User) *Message   // устанавливает отправителя
func (m *Message) Received() *Message        // устанавливает CreatedAt = time.Now()
func (m *Message) Updated() *Message         // устанавливает UpdatedAt = time.Now()
```

## Аутентификация

Файл: `internal/domains/auth.go`

```go
type LoginDTO struct {
    Login    string `json:"login"`
    Password string `json:"password"`
}

type RegisterDTO struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type TokensDTO struct {
    AccessToken  string `json:"accessToken"`
    RefreshToken string `json:"refreshToken"`
}

type ForgetPasswordDTO struct {
    NewPassword string `json:"newPassword"`
}

type ChangePasswordDTO struct {
    NewPassword string `json:"newPassword"`
    OldPassword string `json:"oldPassword"`
}

type SendVerifyEmailMessageDTO struct {
    Email string `json:"email"`
}

type SendForgetPasswordMessageDTO struct {
    Email string `json:"email"`
}
```

`LoginDTO.Login` принимает как email, так и username — валидация на уровне хендлера проверяет оба варианта.

## Пагинация

Файл: `internal/domains/pagination.go`

```go
type Pagination struct {
    Limit  *uint64 `json:"limit,omitempty"`
    Offset *uint64 `json:"offset,omitempty"`
}
```

Оба поля опциональны. Передаётся в запросах к спискам чатов, сообщений и поиска пользователей.

## Настройки и тема

Файл: `internal/domains/setings.go`

```go
type ThemeType int

const (
    ThemeLight ThemeType = iota  // 0
    ThemeDark                    // 1
)

type Settings struct {
    Theme ThemeType `json:"theme"`
}
```

`ThemeLight` является темой по умолчанию. Настройки сохраняются локально в файле `settings.json`.
