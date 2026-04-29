package forgetPassword

import (
	"context"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/security"
	"github.com/DKhorkov/libs/validation"
)

const (
	title = "Сброс пароля"

	width  = 300
	height = 300

	forgetPasswordTokenEntryText = "Код для сброса пароля" //nolint:gosec // наименование переменной
	newPasswordEntryText         = "Пароль"                //nolint:gosec // наименование переменной
	confirmNewPasswordEntryText  = "Подтверждение пароля"  //nolint:gosec // наименование переменной

	forgetPasswordButtonName = "Сбросить пароль" //nolint:gosec // наименование переменной

	passwordChangedText = "Пароль был успешно сброшен.\n\n" +
		"Теперь вы можете авторизоваться под новым паролем!"
)

type Window struct {
	app    fyne.App
	window fyne.Window

	informationWindow interfaces.Window
	useCases          interfaces.UseCases
	validationConfig  config.ValidationConfig
	errorsMapper      interfaces.ErrorsMapper
}

func New(
	app fyne.App,
	informationWindow interfaces.Window,
	useCases interfaces.UseCases,
	validationConfig config.ValidationConfig,
	errorsMapper interfaces.ErrorsMapper,
) *Window {
	return &Window{
		app:               app,
		informationWindow: informationWindow,
		useCases:          useCases,
		validationConfig:  validationConfig,
		errorsMapper:      errorsMapper,
	}
}

func (w *Window) Build(info fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	forgetPasswordTokenEntry := widget.NewEntry()
	forgetPasswordTokenEntry.SetPlaceHolder(forgetPasswordTokenEntryText)

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder(newPasswordEntryText)

	confirmNewPasswordEntry := widget.NewPasswordEntry()
	confirmNewPasswordEntry.SetPlaceHolder(confirmNewPasswordEntryText)

	forgetPasswordButton := widget.NewButton(forgetPasswordButtonName, func() {
		bytesUserID, err := security.RawDecode(forgetPasswordTokenEntry.Text)
		if err != nil {
			err = w.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
			dialog.ShowError(err, w.window)

			return
		}

		if _, err = strconv.ParseUint(string(bytesUserID), 10, 64); err != nil {
			err = w.errorsMapper.Map(errors.ErrInvalidForgetPasswordToken)
			dialog.ShowError(err, w.window)

			return
		}

		if !validation.ValidateValueByRules(
			newPasswordEntry.Text,
			w.validationConfig.PasswordRegExps,
		) {
			err = w.errorsMapper.Map(errors.ErrInvalidPassword)
			dialog.ShowError(err, w.window)

			return
		}

		if newPasswordEntry.Text != confirmNewPasswordEntry.Text {
			err = w.errorsMapper.Map(errors.ErrPasswordDoesNotMatch)
			dialog.ShowError(err, w.window)

			return
		}

		go func() {
			ctx := context.Background()

			if err = w.useCases.ForgetPassword(
				ctx,
				forgetPasswordTokenEntry.Text,
				newPasswordEntry.Text,
			); err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				w.informationWindow.Build(widget.NewLabel(passwordChangedText))
				w.informationWindow.Show()

				w.window.Close()
			})
		}()
	})

	content := container.NewVBox(
		info,
		forgetPasswordTokenEntry,
		newPasswordEntry,
		confirmNewPasswordEntry,
		forgetPasswordButton,
	)

	window.SetContent(content)
	window.CenterOnScreen()

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
