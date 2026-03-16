package auth

import (
	"context"
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
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
	passwordEntryText        = "Пароль"
	usernameEntryText        = "Имя пользователя (от 5 символов)"
	confirmPasswordEntryText = "Подтверждение пароля"

	loginTabIndex    = 0
	registerTabIndex = 1
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases   interfaces.UseCases
	chatWindow interfaces.Window
}

func New(
	app fyne.App,
	chatWindow interfaces.Window,
	useCases interfaces.UseCases,
) *Window {
	return &Window{
		app:        app,
		useCases:   useCases,
		chatWindow: chatWindow,
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
	w.window.Show()
}

func (w *Window) Close() {
	w.window.Close()
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
		if loginEmailEntry.Text == "" || loginPasswordEntry.Text == "" {
			dialog.ShowError(errors.New("заполните все поля"), w.window)

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
		if registerEmailEntry.Text == "" || registerUsernameEntry.Text == "" ||
			registerPasswordEntry.Text == "" || registerConfirmPasswordEntry.Text == "" {
			dialog.ShowError(errors.New("заполните все поля"), w.window)

			return
		}

		if registerPasswordEntry.Text != registerConfirmPasswordEntry.Text {
			dialog.ShowError(errors.New("пароли не совпадают"), w.window)

			return
		}

		if len(registerUsernameEntry.Text) < 5 {
			dialog.ShowError(errors.New("имя пользователя должно быть не менее 5 символов"), w.window)

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

					dialog.ShowError(errors.New("Ошибка регистрации: "+err.Error()), w.window)
				})

				return
			}

			fyne.Do(func() {
				progressBar.Hidden = true

				dialog.ShowInformation("Успешная регистрация", "Регистрация успешна! Теперь войдите.", w.window)

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
