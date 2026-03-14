package searchUsers

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

const (
	title = "Поиск пользователей"

	width  = 400
	height = 500

	searchEntryText = "Введите имя пользователя..."

	usernameLabelText   = "Имя пользователя"
	emailLabelText      = "email"
	resultsLabelText    = "Результаты:"
	searchLabelText     = "Поиск..."
	usersFoundLabelText = "Найдено пользователей: %d"

	usernameLabelIndex = 1
	emailLabelIndex    = 2

	limit  = 0
	offset = 0
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases    interfaces.UseCases
	errorWindow interfaces.Window
}

func New(app fyne.App, useCases interfaces.UseCases, errorWindow interfaces.Window) *Window {
	return &Window{app: app, useCases: useCases, errorWindow: errorWindow}
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

	// Оставляем пустым сначала, чтобы ничего не отображать
	statusLabel := widget.NewLabel("")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(searchEntryText)
	searchEntry.OnSubmitted = func(username string) {
		if username == "" {
			return
		}

		statusLabel.SetText(searchLabelText)

		go func() {
			ctx := context.Background()

			users, err = w.useCases.SearchUsers(ctx, username, limit, offset)
			if err != nil {
				fyne.Do(func() {
					w.errorWindow.Build(widget.NewLabel("Ошибка: " + err.Error()))
					w.errorWindow.Show()
				})

				return
			}

			fyne.Do(func() {
				statusLabel.SetText(fmt.Sprintf(usersFoundLabelText, len(users)))

				usersList.Refresh()
			})
		}()
	}

	// Создаем верхнюю панель со всеми элементами управления
	topContent := container.NewVBox(
		widget.NewLabel(title),
		searchEntry,
		statusLabel,
		widget.NewLabel(resultsLabelText),
	)

	// Устанавливаем Border: topContent сверху, usersList в центре (заполнит всё окно)
	window.SetContent(container.NewBorder(
		topContent, // Сверху (занимает минимум места)
		nil,        // Снизу (пусто)
		nil,        // Слева (пусто)
		nil,        // Справа (пусто)
		usersList,  // В центре (РАСТЯГИВАЕТСЯ на всё оставшееся место)
	))

	w.window = window
}

func (w *Window) Show() {
	w.window.Show()
}

func (w *Window) Close() {
	w.window.Close()
}
