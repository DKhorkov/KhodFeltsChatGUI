package errors

import "errors"

var (
	ErrUserNotFound      = errors.New(`user not found`)
	ErrUserAlreadyExists = errors.New(`user already exists`)
	ErrUpdateUsername    = errors.New("pq: duplicate key value violates unique constraint \"users_username_key\"\n")
)
