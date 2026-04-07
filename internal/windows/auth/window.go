package auth

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
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
	forgetPasswordButtonName  = "Сбросить пароль" //nolint:gosec // наименование переменной

	emailEntryText           = "Почта"
	passwordEntryText        = "Пароль" //nolint:gosec // наименование переменной
	usernameEntryText        = "Логин"
	confirmPasswordEntryText = "Подтверждение пароля" //nolint:gosec // наименование переменной

	loginTabIndex    = 0
	registerTabIndex = 1

	successRegisterTitle = "Успешная регистрация"
	successRegisterText  = "Регистрация успешна! Теперь войдите."
	verifyEmailSentTitle = "Письмо подтверждения отправлено"
	verifyEmailSentText  = "Письмо для подтверждения почты было отправлено по адресу <%s>.\n\n" +
		"Пожалуйста, перейдите по ссылке из письма для подтверждения почты."
	forgetPasswordText = "Письмо с кодом для сброса пароля было отправлено по адресу <%s>.\n\n" +
		"Пожалуйста, используйте полученный код и укажите новый пароль."
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases                         interfaces.UseCases
	chatWindow, forgetPasswordWindow interfaces.Window
	validationConfig                 config.ValidationConfig
	errorsMapper                     interfaces.ErrorsMapper
}

func New(
	app fyne.App,
	chatWindow, forgetPasswordWindow interfaces.Window,
	useCases interfaces.UseCases,
	validationConfig config.ValidationConfig,
	errorsMapper interfaces.ErrorsMapper,
) *Window {
	return &Window{
		app:                  app,
		useCases:             useCases,
		chatWindow:           chatWindow,
		forgetPasswordWindow: forgetPasswordWindow,
		validationConfig:     validationConfig,
		errorsMapper:         errorsMapper,
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
			err := w.errorsMapper.Map(errors.ErrInvalidEmail)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			loginPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			err := w.errorsMapper.Map(errors.ErrInvalidPassword)
			dialog.ShowError(err, w.window)

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
					err = w.errorsMapper.Map(errors.ErrLogin)
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
				err := w.errorsMapper.Map(errors.ErrInvalidEmail)
				dialog.ShowError(err, w.window)

				return
			}

			go func() {
				ctx := context.Background()

				if err := w.useCases.SendVerifyEmailMessage(ctx, loginEmailEntry.Text); err != nil {
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

	forgetPasswordButton := widget.NewButtonWithIcon(
		forgetPasswordButtonName,
		theme.DeleteIcon(),
		func() {
			if !validation.ValidateValueByRule(
				loginEmailEntry.Text,
				w.validationConfig.EmailRegExp,
			) {
				err := w.errorsMapper.Map(errors.ErrInvalidEmail)
				dialog.ShowError(err, w.window)

				return
			}

			go func() {
				ctx := context.Background()

				if err := w.useCases.SendForgetPasswordMessage(
					ctx,
					loginEmailEntry.Text,
				); err != nil {
					fyne.Do(func() {
						dialog.ShowError(err, w.window)
					})

					return
				}

				fyne.Do(func() {
					w.forgetPasswordWindow.Build(
						widget.NewLabel(
							fmt.Sprintf(
								forgetPasswordText,
								loginEmailEntry.Text,
							),
						),
					)
					w.forgetPasswordWindow.Show()
				})
			}()
		})
	forgetPasswordButton.Importance = widget.DangerImportance

	loginTab := container.NewVBox(
		widget.NewLabelWithStyle(loginTabName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		loginEmailEntry,
		loginPasswordEntry,
		loginButton,
		progressBar,
		sendVerifyEmailButton,
		forgetPasswordButton,
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

	registerButton := widget.NewButton(registerButtonName, func() {
		if !validation.ValidateValueByRule(
			registerEmailEntry.Text,
			w.validationConfig.EmailRegExp,
		) {
			err := w.errorsMapper.Map(errors.ErrInvalidEmail)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerUsernameEntry.Text,
			w.validationConfig.UsernameRegExps,
		) {
			err := w.errorsMapper.Map(errors.ErrInvalidUsername)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			err := w.errorsMapper.Map(errors.ErrInvalidPassword)
			dialog.ShowError(err, w.window)

			return
		}

		if registerPasswordEntry.Text != registerConfirmPasswordEntry.Text {
			err := w.errorsMapper.Map(errors.ErrPasswordDoesNotMatch)
			dialog.ShowError(err, w.window)

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
					err = w.errorsMapper.Map(errors.ErrLogin)
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
		registerButton,
	)

	// Обновляем содержимое вкладок
	tabs.Items[loginTabIndex].Content = loginTab
	tabs.Items[registerTabIndex].Content = registerTab

	return tabs
}
