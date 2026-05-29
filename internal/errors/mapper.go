package errors

import (
	"errors"
	"sort"
	"strings"
)

var (
	// Users.
	errUserNotFound     = errors.New("Такого пользователя не существует")
	errUserAlreadyExist = errors.New("Пользователь с такой почтой или логином уже существует")
	errUpdateUsername   = errors.New("Пользователь с таким логином уже существует")

	// Auth.
	errEmailAlreadyExist                      = errors.New("Этот почтовый адрес уже занят")
	errLoginFailed                            = errors.New("Неверный логин или пароль")
	errEmailNotConfirmed                      = errors.New("Почта не была подтверждена")
	errEmailAlreadyConfirmed                  = errors.New("Эта почта уже подтверждена")
	errAccessTokenDoesNotBelongToRefreshToken = errors.New("Ошибка авторизации")
	errRegister                               = errors.New("Ошибка регистрации")
	errInvalidJwtToken                        = errors.New("Ошибка авторизации")
	errValidationFailed                       = errors.New("Ошибка авторизации")
	errInvalidPassword                        = errors.New(
		"Пароль должен быть на латинице, не менее 8 символов в длину и содержать как минимум одну букву" +
			" в верхнем и нижнем регистре, цифру и спецсимвол",
	)
	errInvalidLogin = errors.New(
		"Некорректный email или логин. " +
			"Логин должен быть не менее 5 символов в длину и содержать только латинские буквы и цифры",
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
	errTokenExpired = errors.New("Токен устарел")

	// Chats.
	errInvalidChat         = errors.New("Неверный чат")
	errUserIsNotChatMember = errors.New("У вас нет доступа к этому чату")
	errChatNotFound        = errors.New("Чат не найден")
	errChatAlreadyExist    = errors.New("Такой чат уже существует")
	errGetUserChats        = errors.New(`Не удалось получить чаты пользователя`)
	errCreateChat          = errors.New(`Не удалось создать чат`)
	errGetChatMessages     = errors.New(`Не удалось получить сообщения для чата`)

	// Messages.
	errMessageNotFound  = errors.New("Сообщение не найдено")
	errNotMessageAuthor = errors.New("Только автор сообщения может выполнить это действие")

	// File Storage.
	errInvalidImageFormat = errors.New("Неподдерживаемый формат изображения. Допустимые форматы: JPEG, PNG, WebP, GIF")
	errFileTooLarge       = errors.New("Файл слишком большой")
	errFileNotFound       = errors.New("Файл не найден")

	// Settings.
	errSettingsNotFound            = errors.New("Не удалось получить настройки")
	errWebPushSubscriptionNotFound = errors.New("Подписка на push-уведомления не найдена")
	errWebPushSubscriptionExpired  = errors.New("Подписка на push-уведомления устарела")

	// Limit.
	errLimitExceeded = errors.New("Превышен лимит. Попробуйте позже")

	// ErrDefault - дефолтная ошибка для случаев, когда маппинг не произошел.
	ErrDefault = errors.New("Что-то пошло не так...")
)

var rawMapping = map[string]error{
	// Users
	ErrUserNotFound.Error():      errUserNotFound,
	ErrUserAlreadyExists.Error(): errUserAlreadyExist,
	ErrUpdateUsername.Error():    errUpdateUsername,

	// Auth
	ErrEmailAlreadyConfirmed.Error():                  errEmailAlreadyConfirmed,
	ErrEmailNotConfirmed.Error():                      errEmailNotConfirmed,
	ErrEmailAlreadyExists.Error():                     errEmailAlreadyExist,
	ErrLogin.Error():                                  errLoginFailed,
	ErrWrongPassword.Error():                          errLoginFailed,
	ErrAccessTokenDoesNotBelongToRefreshToken.Error(): errAccessTokenDoesNotBelongToRefreshToken,
	ErrInvalidJwtToken.Error():                        errInvalidJwtToken,
	ErrNewPasswordEqualToOldPassword.Error():          errNewPasswordEqualToOldPassword,
	ErrValidationFailed.Error():                       errValidationFailed,
	ErrInvalidLogin.Error():                           errInvalidLogin,
	ErrInvalidPassword.Error():                        errInvalidPassword,
	ErrInvalidUsername.Error():                        errInvalidUsername,
	ErrInvalidEmail.Error():                           errInvalidEmail,
	ErrPasswordDoesNotMatch.Error():                   errPasswordDoesNotMatch,
	ErrRegister.Error():                               errRegister,
	ErrWebsocket.Error():                              ErrDefault,
	ErrWebsocketClosed.Error():                        ErrDefault,
	ErrInvalidForgetPasswordToken.Error():             errInvalidForgetPasswordToken,
	ErrTokenExpired.Error():                           errTokenExpired,

	// Chats
	ErrInvalidChat.Error():         errInvalidChat,
	ErrUserIsNotChatMember.Error(): errUserIsNotChatMember,
	ErrChatNotFound.Error():        errChatNotFound,
	ErrChatAlreadyExists.Error():   errChatAlreadyExist,
	ErrGetUserChats.Error():        errGetUserChats,
	ErrCreateChat.Error():          errCreateChat,
	ErrGetChatMessages.Error():     errGetChatMessages,

	// Messages
	ErrNotMessageAuthor.Error(): errNotMessageAuthor,
	ErrMessageNotFound.Error():  errMessageNotFound,

	// File Storage
	ErrInvalidImageFormat.Error(): errInvalidImageFormat,
	ErrFileTooLarge.Error():       errFileTooLarge,
	ErrFileNotFound.Error():       errFileNotFound,

	// Settings
	ErrSettingsNotFound.Error():            errSettingsNotFound,
	ErrWebPushSubscriptionNotFound.Error(): errWebPushSubscriptionNotFound,
	ErrWebPushSubscriptionExpired.Error():  errWebPushSubscriptionExpired,

	// Limit
	ErrLimitExceeded.Error(): errLimitExceeded,
}

type mappingEntry struct {
	key string
	val error
}

type Mapper struct {
	mapping []mappingEntry
}

func New() *Mapper {
	return &Mapper{
		mapping: buildSortedMapping(rawMapping),
	}
}

func (m *Mapper) Map(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	for _, entry := range m.mapping {
		if strings.Contains(errMsg, entry.key) {
			return entry.val
		}
	}

	return ErrDefault
}

func buildSortedMapping(m map[string]error) []mappingEntry {
	entries := make([]mappingEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, mappingEntry{key: k, val: v})
	}

	// Сортировка по убыванию длины ключа, чтобы более специфичные ключи проверялись первыми.
	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].key) > len(entries[j].key)
	})

	return entries
}
