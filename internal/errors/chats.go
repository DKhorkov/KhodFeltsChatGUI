package errors

import "errors"

var (
	ErrCreateChat          = errors.New(`failed to create chat`)
	ErrGetUserChats        = errors.New(`failed to get user chats`)
	ErrGetChatMessages     = errors.New(`failed to get chat messages`)
	ErrInvalidChat         = errors.New(`invalid chat`)
	ErrUserIsNotChatMember = errors.New(`user is not a chat member`)
	ErrChatNotFound        = errors.New(`chat not found`)
	ErrChatAlreadyExists   = errors.New(`chat already exists`)
)
