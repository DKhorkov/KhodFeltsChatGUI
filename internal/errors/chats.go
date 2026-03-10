package errors

import "errors"

var (
	ErrGetUserChats    = errors.New(`failed to get user chats`)
	ErrCreateChat      = errors.New(`failed to create chat`)
	ErrGetChatMessages = errors.New(`failed to get chat messages`)
)
