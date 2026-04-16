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
	errEmailAlreadyExist                      = errors.New("Этот почтовый адрес уже занят")
	errLoginFailed                            = errors.New("Неверный логин или пароль")
	errEmailNotConfirmed                      = errors.New("Почта не была подтвержён")
	errEmailAlreadyConfirmed                  = errors.New("Эта почта уже подтверждена")
	errAccessTokenDoesNotBelongToRefreshToken = errors.New("Ошибка авторизации")
	errRegister                               = errors.New("Ошибка регистрации")
	errInvalidJwtToken                        = errors.New("Ошибка авторизации")
	errValidationFailed                       = errors.New("Ошибка авторизации")
	errInvalidPassword                        = errors.New(
		"Пароль должен быть на латинице, не менее 8 символов в длину и содержать как минимум одну букву" +
			" в верхнем и нижнем регистре, цифру и спецсимвол",
	)
	errInvalidUsername = errors.New(
		"Логин должен быть не менее 5 символов в длину и содержать только латинские буквы и цифры",
	)
	errInvalidEmail                  = errors.New("Некорректный адрес электронной почты")
	errPasswordDoesNotMatch          = errors.New("Пароли не совпадают")
	errInvalidForgetPasswordToken    = errors.New("Некорректный код для сброса пароля")
	errNewPasswordEqualToOldPassword = errors.New(
		"Старый пароль не может быть использован в качестве нового пароля",
	)

	// Chats.
	errInvalidChat         = errors.New("Неверный чат")
	errUserIsNotChatMember = errors.New("У вас нет доступа к этому чату")
	errChatNotFound        = errors.New("Чат не найден")
	errChatAlreadyExist    = errors.New("Такой чат уже существует")
	errGetUserChats        = errors.New(`Не удалось получить чаты пользователя`)
	errCreateChat          = errors.New(`Не удалось создать чат`)
	errGetChatMessages     = errors.New(`Не удалось получить сообщения для чата`)

	// Limit.
	errLimitExceeded = errors.New("Превышен лимит. Попробуйте позже")

	// Default.
	errDefault = errors.New("Что-то пошло не так...")
)

var mapping = map[string]error{
	// Users
	ErrUserNotFound.Error():      errUserNotFound,
	ErrUserAlreadyExists.Error(): errUserAlreadyExist,

	// Auth
	ErrEmailAlreadyConfirmed.Error():                  errEmailAlreadyConfirmed,
	ErrEmailNotConfirmed.Error():                      errEmailNotConfirmed,
	ErrEmailAlreadyExists.Error():                     errEmailAlreadyExist,
	ErrLogin.Error():                                  errLoginFailed,
	ErrWrongPassword.Error():                          errLoginFailed,
	ErrAccessTokenDoesNotBelongToRefreshToken.Error(): errAccessTokenDoesNotBelongToRefreshToken,
	ErrInvalidJwtToken.Error():                        errInvalidJwtToken,
	ErrValidationFailed.Error():                       errValidationFailed,
	ErrInvalidPassword.Error():                        errInvalidPassword,
	ErrInvalidUsername.Error():                        errInvalidUsername,
	ErrInvalidEmail.Error():                           errInvalidEmail,
	ErrPasswordDoesNotMatch.Error():                   errPasswordDoesNotMatch,
	ErrRegister.Error():                               errRegister,
	ErrWebsocket.Error():                              errDefault,
	ErrWebsocketClosed.Error():                        errDefault,
	ErrInvalidForgetPasswordToken.Error():             errInvalidForgetPasswordToken,
	ErrNewPasswordEqualToOldPassword.Error():          errNewPasswordEqualToOldPassword,

	// Chats
	ErrInvalidChat.Error():         errInvalidChat,
	ErrUserIsNotChatMember.Error(): errUserIsNotChatMember,
	ErrChatNotFound.Error():        errChatNotFound,
	ErrChatAlreadyExists.Error():   errChatAlreadyExist,
	ErrGetUserChats.Error():        errGetUserChats,
	ErrCreateChat.Error():          errCreateChat,
	ErrGetChatMessages.Error():     errGetChatMessages,

	// Limit
	ErrLimitExceeded.Error(): errLimitExceeded,
}

func New() *Mapper {
	return &Mapper{
		mapping: mapping,
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
