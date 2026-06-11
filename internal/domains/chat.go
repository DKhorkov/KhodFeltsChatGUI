package domains

import (
	"slices"
	"time"
)

type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"

	minMembersCount         = 1 // группа на одного
	privateChatMembersCount = 2
)

var ChatTypes = []ChatType{
	ChatTypePrivate,
	ChatTypeGroup,
}

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

func (c *Chat) IsValid() bool {
	if !slices.Contains(ChatTypes, c.Type) {
		return false
	}

	if len(c.Members) < minMembersCount {
		return false
	}

	if c.Type == ChatTypePrivate && len(c.Members) != privateChatMembersCount {
		return false
	}

	return true
}

type CreateChatDTO struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Type        ChatType `json:"type"`
	MemberIDs   []uint64 `json:"memberIDs,omitempty"`
}

func (c *CreateChatDTO) IsValid() bool {
	if !slices.Contains(ChatTypes, c.Type) {
		return false
	}

	// На бэкенде текущий пользователь автоматически добавляется в чат при создании.
	// Поэтому проверяем, что для приватного чата был выбран именно второй участник
	if c.Type == ChatTypePrivate && len(c.MemberIDs) != minMembersCount {
		return false
	}

	return true
}
