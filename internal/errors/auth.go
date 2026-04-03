package errors

import "errors"

var (
	ErrRegister             = errors.New(`registration failed`)
	ErrLogin                = errors.New(`login failed`)
	ErrLogout               = errors.New(`logout failed`)
	ErrRefreshTokens        = errors.New(`refresh tokens failed`)
	ErrInvalidPassword      = errors.New(`invalid password`)
	ErrInvalidUsername      = errors.New(`invalid username`)
	ErrInvalidEmail         = errors.New(`invalid email`)
	ErrPasswordDoesNotMatch = errors.New(`passwords does not match`)
	ErrInvalidForgetPasswordToken = errors.New(`invalid forget password token`)
)
