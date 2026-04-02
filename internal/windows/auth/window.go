package auth

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/validation"
)

const (
	title = "Вход / Регистрация"

	width  = 500
	height = 450

	loginTabName    = "Вход"
	registerTabName = "Регистрация"

	loginButtonName           = "Войти"
	registerButtonName        = "Зарегистрироваться"
	sendVerifyEmailButtonName = "Отправить повторно письмо для подтверждения почты"

	emailEntryText           = "Почта"
	passwordEntryText        = "Пароль" //nolint:gosec // наименование переменной
	usernameEntryText        = "Логин"
	confirmPasswordEntryText = "Подтверждение пароля" //nolint:gosec // наименование переменной

	loginTabIndex    = 0
	registerTabIndex = 1

	successRegisterTitle = "Успешная регистрация"
	successRegisterText  = "Регистрация успешна! Теперь войдите."
	verifyEmailSentTitle = "Письмо подтверждения отправлено"
	verifyEmailSentText  = "Письмо для подтверждения почты было успешно отправлено по адресу <%s>.\n\n" +
		"Пожалуйста, перейдите по ссылке из письма для подтверждения почты."
)

var (
	errInvalidPassword = errors.New(
		"пароль должен быть на латинице, не менее 8 символов в длину и содержать как минимум одну букву" +
			" в верхнем и нижнем регистре, цифру и спецсимвол",
	)
	errInvalidUsername = errors.New(
		"имя пользователя должно быть не менее 5 символов в длину и содержать только латинские буквы и цифры",
	)
	errInvalidEmail         = errors.New("некорректный адрес электронной почты")
	errPasswordDoesNotMatch = errors.New("пароли не совпадают")
	errRegistration         = errors.New("ошибка регистрации")
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases         interfaces.UseCases
	chatWindow       interfaces.Window
	validationConfig config.ValidationConfig
}

func New(
	app fyne.App,
	chatWindow interfaces.Window,
	useCases interfaces.UseCases,
	validationConfig config.ValidationConfig,
) *Window {
	return &Window{
		app:              app,
		useCases:         useCases,
		chatWindow:       chatWindow,
		validationConfig: validationConfig,
	}
}

func (w *Window) SetChatWindow(chatWindow interfaces.Window) {
	w.chatWindow = chatWindow
}

func (w *Window) Build(_ fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	// Устанавливаем обработчик закрытия окна авторизации
	window.SetCloseIntercept(func() {
		w.app.Quit()
	})

	tabs := w.buildTabs()

	window.SetContent(tabs)

	w.window = window
}

func (w *Window) Show() {
	if w.window == nil {
		return
	}

	w.window.Show()
}

func (w *Window) Close() {
	if w.window == nil {
		return
	}

	w.window.Close()
	w.window = nil
}

func (w *Window) buildTabs() *container.AppTabs {
	// Создаем вкладки пустыми для правильной настройки кнопок регистрации и логина
	tabs := container.NewAppTabs(
		container.NewTabItem(loginTabName, nil),
		container.NewTabItem(registerTabName, nil),
	)

	// Вход
	loginEmailEntry := widget.NewEntry()
	loginEmailEntry.SetPlaceHolder(emailEntryText)

	loginPasswordEntry := widget.NewPasswordEntry()
	loginPasswordEntry.SetPlaceHolder(passwordEntryText)

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hidden = true

	loginButton := widget.NewButton(loginButtonName, func() {
		if !validation.ValidateValueByRule(loginEmailEntry.Text, w.validationConfig.EmailRegExp) {
			dialog.ShowError(errInvalidEmail, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			loginPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			dialog.ShowError(errInvalidPassword, w.window)

			return
		}

		progressBar.Hidden = false

		go func() {
			ctx := context.Background()

			if _, err := w.useCases.Login(
				ctx,
				loginEmailEntry.Text,
				loginPasswordEntry.Text,
			); err != nil {
				fyne.Do(func() {
					progressBar.Hidden = true

					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				w.window.Close()

				w.chatWindow.Build(nil)
				w.chatWindow.Show()
			})
		}()
	})

	sendVerifyEmailButton := widget.NewButtonWithIcon(
		sendVerifyEmailButtonName,
		theme.MailComposeIcon(),
		func() {
			if !validation.ValidateValueByRule(
				loginEmailEntry.Text,
				w.validationConfig.EmailRegExp,
			) {
				dialog.ShowError(errInvalidEmail, w.window)

				return
			}

			go func() {
				ctx := context.Background()

				if err := w.useCases.SendVerifyEmail(ctx, loginEmailEntry.Text); err != nil {
					fyne.Do(func() {
						dialog.ShowError(err, w.window)
					})

					return
				}

				fyne.Do(func() {
					dialog.ShowInformation(
						verifyEmailSentTitle,
						fmt.Sprintf(verifyEmailSentText, loginEmailEntry.Text),
						w.window,
					)
				})
			}()
		})
	sendVerifyEmailButton.Importance = widget.WarningImportance

	loginTab := container.NewVBox(
		widget.NewLabelWithStyle(loginTabName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		loginEmailEntry,
		loginPasswordEntry,
		loginButton,
		progressBar,
		sendVerifyEmailButton,
	)

	// Регистрация
	registerEmailEntry := widget.NewEntry()
	registerEmailEntry.SetPlaceHolder(emailEntryText)

	registerUsernameEntry := widget.NewEntry()
	registerUsernameEntry.SetPlaceHolder(usernameEntryText)

	registerPasswordEntry := widget.NewPasswordEntry()
	registerPasswordEntry.SetPlaceHolder(passwordEntryText)

	registerConfirmPasswordEntry := widget.NewPasswordEntry()
	registerConfirmPasswordEntry.SetPlaceHolder(confirmPasswordEntryText)

	registerBtn := widget.NewButton(registerButtonName, func() {
		if !validation.ValidateValueByRule(
			registerEmailEntry.Text,
			w.validationConfig.EmailRegExp,
		) {
			dialog.ShowError(errInvalidEmail, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerUsernameEntry.Text,
			w.validationConfig.UsernameRegExps,
		) {
			dialog.ShowError(errInvalidUsername, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			dialog.ShowError(errInvalidPassword, w.window)

			return
		}

		if registerPasswordEntry.Text != registerConfirmPasswordEntry.Text {
			dialog.ShowError(errPasswordDoesNotMatch, w.window)

			return
		}

		progressBar.Hidden = false

		go func() {
			ctx := context.Background()

			registerData := domains.RegisterDTO{
				Username: registerUsernameEntry.Text,
				Password: registerPasswordEntry.Text,
				Email:    registerEmailEntry.Text,
			}

			if _, err := w.useCases.Register(ctx, registerData); err != nil {
				fyne.Do(func() {
					progressBar.Hidden = true

					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				progressBar.Hidden = true

				dialog.ShowInformation(
					successRegisterTitle,
					successRegisterText,
					w.window,
				)

				// Заполняем поля входа тем же email и паролем
				loginEmailEntry.SetText(registerEmailEntry.Text)
				loginPasswordEntry.SetText(registerPasswordEntry.Text)

				// Очищаем поля регистрации
				registerEmailEntry.SetText("")
				registerUsernameEntry.SetText("")
				registerPasswordEntry.SetText("")
				registerConfirmPasswordEntry.SetText("")

				// Переключаемся на вкладку входа
				tabs.SelectIndex(loginTabIndex)
			})
		}()
	})

	registerTab := container.NewVBox(
		widget.NewLabelWithStyle(registerTabName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		registerEmailEntry,
		registerUsernameEntry,
		registerPasswordEntry,
		registerConfirmPasswordEntry,
		registerBtn,
	)

	// Обновляем содержимое вкладок
	tabs.Items[loginTabIndex].Content = loginTab
	tabs.Items[registerTabIndex].Content = registerTab

	return tabs
}
