package errors

import (
	"errors"
	"strings"
)

type Mapper struct {
	mapping map[string]error
}

var defaultError = errors.New("Что-то пошло не так...")

func New() *Mapper {
	return &Mapper{
		mapping: map[string]error{
			"user not found":                                errors.New("Такого пользователя не суествует"),
			"user already exists":                           errors.New("Такой пользователь уже существует"),
			"email not confirmed":                           errors.New("Почтовый адрес не был подтвержён"),
			"email already confirmed":                       errors.New("Этот почтовый адрес уже занят"),
			"wrong password":                                errors.New("Неверный логин или пароль"),
			"access token does not belong to refresh token": errors.New("Ошибка авторизации"),
			"invalid chat":                                  errors.New("Неверный чат"),
			"user is not a chat member":                     errors.New("У вас нет доступа к этому чату"),
			"chat not found":                                errors.New("Чат не найден"),
			"chat already exists":                           errors.New("Такой чат уже существует"),
			"invalid jwt token":                             errors.New("Ошибка авторизации"),
			"validation failed":                             errors.New("Ошибка авторизации"),
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

	return defaultError
}
