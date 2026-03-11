package errors

import "errors"

var (
	ErrWebsocket       = errors.New("websocket error")
	ErrWebsocketClosed = errors.New("websocket closed")
)
