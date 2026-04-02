package errors

import (
	"errors"
	"strings"
)

type Mapper struct {
	mapping map[string]error
}

var (
	errUserNotFound                           = errors.New("Такого пользователя не суествует")
	errUserAlreadyExist                       = errors.New("Такой пользователь уже существует")
	errEmailNotConfirmed                      = errors.New("Почтовый адрес не был подтвержён")
	errEmailAlreadyExist                      = errors.New("Этот почтовый адрес уже занят")
	errFailedLogin                            = errors.New("Неверный логин или пароль")
	errAccessTokenDoesNotBelongToRefreshToken = errors.New("Ошибка авторизации")
	errInvalidChat                            = errors.New("Неверный чат")
	errUserNotIsMember                        = errors.New("У вас нет доступа к этому чату")
	errChatNotFound                           = errors.New("Чат не найден")
	errChatAlreadyExist                       = errors.New("Такой чат уже существует")
	errInvalidJwtToken                        = errors.New("Ошибка авторизации")
	errValidationFailed                       = errors.New("Ошибка авторизации")
	errDefault                                = errors.New("Что-то пошло не так...")
	errInvalidPassword                        = errors.New(
		"Пароль должен быть на латинице, не менее 8 символов в длину и содержать как минимум одну букву" +
			" в верхнем и нижнем регистре, цифру и спецсимвол",
	)
	errInvalidUsername = errors.New(
		"Имя пользователя должно быть не менее 5 символов в длину и содержать только латинские буквы и цифры",
	)
	errInvalidEmail         = errors.New("Некорректный email")
	errPasswordDoesNotMatch = errors.New("Пароли не совпадают")
	errRegister             = errors.New("Ошибка регистрации")
	errGetUserChats         = errors.New(`Не удалось получить чаты пользователя`)
	errCreateChat           = errors.New(`Не удалось создать чат`)
	errGetChatMessages      = errors.New(`Не удалось получить сообщения для чата`)
)

func New() *Mapper {
	return &Mapper{
		mapping: map[string]error{
			ErrUserNotFound.Error():                         errUserNotFound,
			"user already exists":                           errUserAlreadyExist,
			"email not confirmed":                           errEmailNotConfirmed,
			"email already confirmed":                       errEmailAlreadyExist,
			ErrLogin.Error():                                errFailedLogin,
			"access token does not belong to refresh token": errAccessTokenDoesNotBelongToRefreshToken,
			"invalid chat":                                  errInvalidChat,
			"user is not a chat member":                     errUserNotIsMember,
			"chat not found":                                errChatNotFound,
			"chat already exists":                           errChatAlreadyExist,
			"invalid jwt token":                             errInvalidJwtToken,
			"validation failed":                             errValidationFailed,
			ErrInvalidPassword.Error():                      errInvalidPassword,
			ErrInvalidUsername.Error():                      errInvalidUsername,
			ErrInvalidEmail.Error():                         errInvalidEmail,
			ErrPasswordDoesNotMatch.Error():                 errPasswordDoesNotMatch,
			ErrRegister.Error():                             errRegister,
			ErrGetUserChats.Error():                         errGetUserChats,
			ErrCreateChat.Error():                           errCreateChat,
			ErrGetChatMessages.Error():                      errGetChatMessages,
			ErrWebsocket.Error():                            errDefault,
			ErrWebsocketClosed.Error():                      errDefault,
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
