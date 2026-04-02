package errors

import (
	"errors"
	"strings"
)

type Mapper struct {
	mapping map[string]error
}

var (
	// Users.
	errUserNotFound     = errors.New("Такого пользователя не существует")
	errUserAlreadyExist = errors.New(
		"Пользователь с такой почтой или логином уже существует",
	)

	// Auth.
	errEmailNotConfirmed                      = errors.New("Почта не была подтвержён")
	errEmailAlreadyConfirmed                  = errors.New("Эта почта уже подтверждена")
	errWrongPassword                          = errors.New("Неверный логин или пароль")
	errAccessTokenDoesNotBelongToRefreshToken = errors.New("Ошибка авторизации")
	errInvalidJwtToken                        = errors.New("Ошибка авторизации")
	errValidationFailed                       = errors.New("Ошибка авторизации")

	// Chats.
	errInvalidChat         = errors.New("Неверный чат")
	errUserIsNotChatMember = errors.New("У вас нет доступа к этому чату")
	errChatNotFound        = errors.New("Чат не найден")
	errChatAlreadyExist    = errors.New("Такой чат уже существует")

	// Default.
	errDefault = errors.New("Что-то пошло не так...")
)

func New() *Mapper {
	return &Mapper{
		mapping: map[string]error{
			"user not found":                                errUserNotFound,
			"user already exists":                           errUserAlreadyExist,
			"email not confirmed":                           errEmailNotConfirmed,
			"email already confirmed":                       errEmailAlreadyConfirmed,
			"wrong password":                                errWrongPassword,
			"access token does not belong to refresh token": errAccessTokenDoesNotBelongToRefreshToken,
			"invalid chat":                                  errInvalidChat,
			"user is not a chat member":                     errUserIsNotChatMember,
			"chat not found":                                errChatNotFound,
			"chat already exists":                           errChatAlreadyExist,
			"invalid jwt token":                             errInvalidJwtToken,
			"validation failed":                             errValidationFailed,
		},
	}
}

func (m *Mapper) Map(err error) error {
	if err == nil {
		return nil
	}

	for k, v := range m.mapping {
		if strings.Contains(err.Error(), k) {
			return v
		}
	}

	return errDefault
}
