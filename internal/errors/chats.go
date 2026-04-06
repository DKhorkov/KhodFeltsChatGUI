package errors

import "errors"

var (
	ErrCreateChat      = errors.New(`failed to create chat`)
	ErrUserNotFound    = errors.New(`user not found`)
	ErrGetUserChats    = errors.New(`failed to get user chats`)
	ErrGetChatMessages = errors.New(`failed to get chat messages`)
)
