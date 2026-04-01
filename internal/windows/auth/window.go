package auth

import (
	"context"

	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/validation"
)

const (
	title = "Вход / Регистрация"

	width  = 400
	height = 450

	loginTabName    = "Вход"
	registerTabName = "Регистрация"

	loginButtonName    = "Войти"
	registerButtonName = "Зарегистрироваться"

	emailEntryText           = "Email"
	passwordEntryText        = "Пароль" //nolint:gosec // наименование переменной
	usernameEntryText        = "Имя пользователя (от 5 символов)"
	confirmPasswordEntryText = "Подтверждение пароля" //nolint:gosec // наименование переменной

	loginTabIndex    = 0
	registerTabIndex = 1
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases         interfaces.UseCases
	chatWindow       interfaces.Window
	validationConfig config.ValidationConfig
	errMapper        *customerrors.Mapper
}

func New(
	app fyne.App,
	chatWindow interfaces.Window,
	useCases interfaces.UseCases,
	validationConfig config.ValidationConfig,
	errMapper *customerrors.Mapper,
) *Window {
	return &Window{
		app:              app,
		useCases:         useCases,
		chatWindow:       chatWindow,
		validationConfig: validationConfig,
		errMapper:        errMapper,
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
			err := w.errMapper.Map(customerrors.ErrLogin)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			loginPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			err := w.errMapper.Map(customerrors.ErrLogin)
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

					dialog.ShowError(w.errMapper.Map(err), w.window)
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

	loginTab := container.NewVBox(
		widget.NewLabelWithStyle(loginTabName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		loginEmailEntry,
		loginPasswordEntry,
		loginButton,
		progressBar,
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
			err := w.errMapper.Map(customerrors.ErrInvalidEmail)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerUsernameEntry.Text,
			w.validationConfig.UsernameRegExps,
		) {
			err := w.errMapper.Map(customerrors.ErrInvalidUsername)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			registerPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			err := w.errMapper.Map(customerrors.ErrInvalidPassword)
			dialog.ShowError(err, w.window)

			return
		}

		if registerPasswordEntry.Text != registerConfirmPasswordEntry.Text {
			err := w.errMapper.Map(customerrors.ErrPasswordDoesNotMatch)
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

					err = w.errMapper.Map(customerrors.ErrRegister)
					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				progressBar.Hidden = true

				dialog.ShowInformation(
					"Успешная регистрация",
					"Регистрация успешна! Теперь войдите.",
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
