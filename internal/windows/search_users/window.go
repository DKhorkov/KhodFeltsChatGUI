package searchUsers

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/kfcGUI/internal/windows"
)

const (
	title = "Поиск пользователей"

	width  = 400
	height = 500

	searchEntryText = "Введите имя пользователя..."

	searchButtonName = "Найти"

	usernameLabelText = "Имя пользователя"
	emailLabelText    = "email"

	usernameLabelIndex = 1
	emailLabelIndex    = 2

	limit  = 0
	offset = 0
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases interfaces.UseCases
}

func New(app fyne.App, useCases interfaces.UseCases) *Window {
	return &Window{app: app, useCases: useCases}
}

func (w *Window) Build(_ fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	var (
		users []domains.User
		err   error
	)

	usersList := widget.NewList(
		func() int { return len(users) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel(usernameLabelText),
				widget.NewLabel(emailLabelText),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			user := users[id]

			container_ := item.(*fyne.Container)

			usernameLabel := container_.Objects[usernameLabelIndex].(*widget.Label)
			usernameLabel.SetText(user.Username)

			emailLabel := container_.Objects[emailLabelIndex].(*widget.Label)
			emailLabel.SetText(user.Email)
		},
	)

	// Изначально пустой, потому что пользователь еще не совершал поиск
	foundUsersLabel := widget.NewLabel("")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(searchEntryText)
	searchEntry.OnSubmitted = func(username string) {
		if username == "" {
			return
		}

		go func() {
			ctx := context.Background()

			users, err = w.useCases.SearchUsers(ctx, username, limit, offset)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				foundUsersLabel.SetText(windows.FoundUsersText(len(users)))

				usersList.Refresh()
			})
		}()
	}

	searchBtn := widget.NewButton(searchButtonName, func() {
		username := searchEntry.Text
		if username == "" {
			return
		}

		go func() {
			ctx := context.Background()

			users, err = w.useCases.SearchUsers(ctx, username, limit, offset)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				foundUsersLabel.SetText(windows.FoundUsersText(len(users)))

				usersList.Refresh()
			})
		}()
	})

	// Создаем верхнюю панель со всеми элементами управления
	topContent := container.NewVBox(
		searchEntry,
		searchBtn,
		foundUsersLabel,
	)

	// Устанавливаем Border: topContent сверху, usersList в центре (заполнит всё окно)
	window.SetContent(container.NewBorder(
		topContent,
		nil,
		nil,
		nil,
		usersList,
	))

	w.window = window
}

func (w *Window) Show() {
	w.window.Show()
}

func (w *Window) Close() {
	w.window.Close()
}
