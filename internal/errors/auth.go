package errors

import "errors"

var (
	ErrRegister                               = errors.New(`registration failed`)
	ErrLogin                                  = errors.New(`login failed`)
	ErrLogout                                 = errors.New(`logout failed`)
	ErrRefreshTokens                          = errors.New(`refresh tokens failed`)
	ErrInvalidLogin                           = errors.New(`invalid login`)
	ErrInvalidPassword                        = errors.New(`invalid password`)
	ErrInvalidUsername                        = errors.New(`invalid username`)
	ErrInvalidEmail                           = errors.New(`invalid email`)
	ErrPasswordDoesNotMatch                   = errors.New(`passwords does not match`)
	ErrInvalidForgetPasswordToken             = errors.New(`invalid jwt token: invalid forget_password_token`)
	ErrTokenExpired                           = errors.New("token expired")
	ErrEmailAlreadyConfirmed                  = errors.New(`email already confirmed`)
	ErrEmailNotConfirmed                      = errors.New(`email not confirmed`)
	ErrEmailAlreadyExists                     = errors.New(`email already exist`)
	ErrWrongPassword                          = errors.New(`wrong password`)
	ErrAccessTokenDoesNotBelongToRefreshToken = errors.New(
		`access token does not belong to refresh token`,
	)
	ErrInvalidJwtToken               = errors.New(`invalid jwt token`)
	ErrValidationFailed              = errors.New(`validation failed`)
	ErrNewPasswordEqualToOldPassword = errors.New(
		`validation failed: new password can not be equal to old password`,
	)
)
