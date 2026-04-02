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

	// Default.
	errDefault = errors.New("Что-то пошло не так...")
)

func New() *Mapper {
	return &Mapper{
		mapping: map[string]error{
			// Users
			"user not found":      errUserNotFound,
			"user already exists": errUserAlreadyExist,

			// Auth
			"email not confirmed":                           errEmailNotConfirmed,
			"email already confirmed":                       errEmailAlreadyConfirmed,
			"wrong password":                                errWrongPassword,
			"access token does not belong to refresh token": errAccessTokenDoesNotBelongToRefreshToken,
			"invalid jwt token":                             errInvalidJwtToken,
			"validation failed":                             errValidationFailed,
			ErrInvalidPassword.Error():                      errInvalidPassword,
			ErrInvalidUsername.Error():                      errInvalidUsername,
			ErrInvalidEmail.Error():                         errInvalidEmail,
			ErrPasswordDoesNotMatch.Error():                 errPasswordDoesNotMatch,
			ErrInvalidForgetPasswordToken.Error():           errInvalidForgetPasswordToken,
			"new password equal to old password":            errNewPasswordEqualToOldPassword,

			// Chats
			"invalid chat":              errInvalidChat,
			"user is not a chat member": errUserIsNotChatMember,
			"chat not found":            errChatNotFound,
			"chat already exists":       errChatAlreadyExist,
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
