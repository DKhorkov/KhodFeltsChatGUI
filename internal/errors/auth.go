package errors

import "errors"

var (
	ErrLogin         = errors.New(`login failed`)
	ErrRefreshTokens = errors.New(`refresh tokens failed`)
)
