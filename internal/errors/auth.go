package errors

import "errors"

var (
	ErrRegister      = errors.New(`registration failed`)
	ErrLogin         = errors.New(`login failed`)
	ErrLogout        = errors.New(`logout failed`)
	ErrRefreshTokens = errors.New(`refresh tokens failed`)
)
