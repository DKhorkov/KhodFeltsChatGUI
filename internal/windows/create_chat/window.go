package createChat

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

const (
	title = "Создать чат"

	width  = 450
	height = 400

	searchEntryText = "Введите имя пользователя..."

	searchButtonName     = "Найти"
	createChatButtonName = "Создать"

	chatTypeLabelText = "Тип чата:"
	checkLabelText    = ""
	usernameLabelText = "Имя пользователя"
	emailLabelText    = "email"

	checkedLabelIndex  = 1
	usernameLabelIndex = 2
	emailLabelIndex    = 3

	limit  = 0
	offset = 0
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases         interfaces.UseCases
	refreshChatsFunc func(chat domains.Chat)
}

func New(
	app fyne.App,
	useCases interfaces.UseCases,
	refreshChatsFunc func(chat domains.Chat),
) *Window {
	return &Window{
		app:              app,
		useCases:         useCases,
		refreshChatsFunc: refreshChatsFunc,
	}
}

func (w *Window) SetRefreshChatsFunc(f func(chat domains.Chat)) {
	w.refreshChatsFunc = f
}

func (w *Window) Build(_ fyne.CanvasObject) {
	win := w.app.NewWindow(title)
	win.Resize(fyne.NewSize(width, height))

	availableChatTypes := make([]string, 0, len(domains.ChatTypes))
	for _, chatType := range domains.ChatTypes {
		availableChatTypes = append(availableChatTypes, string(chatType))
	}

	typeEntry := widget.NewSelect(availableChatTypes, nil)
	typeEntry.SetSelected(string(domains.ChatTypePrivate))

	var (
		users         []domains.User
		err           error
		selectedUsers = make(map[uint64]bool)
	)

	usersList := widget.NewList(
		func() int { return len(users) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewCheck(checkLabelText, func(bool) {}),
				widget.NewLabel(usernameLabelText),
				widget.NewLabel(emailLabelText),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container_ := item.(*fyne.Container)
			checkLabel := container_.Objects[checkedLabelIndex].(*widget.Check)
			usernameLabel := container_.Objects[usernameLabelIndex].(*widget.Label)
			emailLabel := container_.Objects[emailLabelIndex].(*widget.Label)

			user := users[id]
			usernameLabel.SetText(user.Username)
			emailLabel.SetText(user.Email)

			checkLabel.SetChecked(selectedUsers[user.ID])

			checkLabel.OnChanged = func(checked bool) {
				if checked {
					selectedUsers[user.ID] = true
				} else {
					delete(selectedUsers, user.ID)
				}
			}
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
					dialog.ShowError(err, win)
				})

				return
			}

			fyne.Do(func() {
				foundUsersLabel.SetText(w.foundUsersText(len(users)))

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
					dialog.ShowError(err, win)
				})

				return
			}

			fyne.Do(func() {
				foundUsersLabel.SetText(w.foundUsersText(len(users)))

				usersList.Refresh()
			})
		}()
	})

	createBtn := widget.NewButton(createChatButtonName, nil)
	createBtn.OnTapped = func() {
		chatType := domains.ChatType(typeEntry.Selected)

		if chatType == domains.ChatTypePrivate && len(selectedUsers) == 0 {
			dialog.ShowError(errors.New("укажите хотя бы одного участника"), w.window)

			return
		}

		createBtn.Disable()

		go func() {
			ctx := context.Background()

			members := make([]domains.User, 0, len(selectedUsers))
			for id := range selectedUsers {
				members = append(members, domains.User{ID: id})
			}

			// TODO доработать длч групповых чатов название и описание - добавить поля ввода
			chat := &domains.Chat{
				Type:    chatType,
				Members: members,
			}

			chat, err := w.useCases.CreateChat(ctx, *chat)
			if err != nil {
				fyne.Do(func() {
					createBtn.Enable()

					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				// Обновляем список чатов
				w.refreshChatsFunc(*chat)

				win.Close()
			})
		}()
	}

	chatTypeBox := container.NewHBox(
		widget.NewLabel(chatTypeLabelText),
		typeEntry,
	)

	topContent := container.NewVBox(
		chatTypeBox,
		searchEntry,
		searchBtn,
		foundUsersLabel,
	)

	bottomContent := container.NewCenter(
		createBtn,
	)

	content := container.NewBorder(
		topContent,
		bottomContent,
		nil,
		nil,
		usersList,
	)

	win.SetContent(content)

	win.Show()
}

func (w *Window) Show() {
	w.window.Show()
}

func (w *Window) Close() {
	w.window.Close()
}

func (w *Window) foundUsersText(usersCount int) string {
	// Исключения для чисел от 11 до 14
	if usersCount%100 >= 11 && usersCount%100 <= 14 {
		return fmt.Sprintf("Найдено %d пользователей:", usersCount)
	}

	switch usersCount % 10 {
	case 1:
		return fmt.Sprintf("Найден %d пользователь:", usersCount)
	case 2, 3, 4:
		return fmt.Sprintf("Найдено %d пользователя:", usersCount)
	default:
		return fmt.Sprintf("Найдено %d пользователей:", usersCount)
	}
}
