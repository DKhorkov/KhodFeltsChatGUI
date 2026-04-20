package createChat

import (
	"context"
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/kfcGUI/internal/v1/windows"
	"github.com/DKhorkov/libs/pointers"
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

var errNoChatMembersProvided = errors.New("укажите хотя бы одного участника")

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
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

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

			filters := &domains.UsersFilters{
				Username: pointers.New(username),
			}

			pagination := &domains.Pagination{
				Limit:  pointers.New[uint64](limit),
				Offset: pointers.New[uint64](offset),
			}

			users, err = w.useCases.SearchUsers(ctx, filters, pagination)
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

	searchButton := widget.NewButton(searchButtonName, func() {
		username := searchEntry.Text
		if username == "" {
			return
		}

		go func() {
			ctx := context.Background()

			filters := &domains.UsersFilters{
				Username: pointers.New(username),
			}

			pagination := &domains.Pagination{
				Limit:  pointers.New[uint64](limit),
				Offset: pointers.New[uint64](offset),
			}

			users, err = w.useCases.SearchUsers(ctx, filters, pagination)
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

	createButton := widget.NewButton(createChatButtonName, nil)
	createButton.OnTapped = func() {
		chatType := domains.ChatType(typeEntry.Selected)

		if chatType == domains.ChatTypePrivate && len(selectedUsers) == 0 {
			dialog.ShowError(errNoChatMembersProvided, w.window)

			return
		}

		createButton.Disable()

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
					createButton.Enable()

					dialog.ShowError(err, w.window)
				})

				return
			}

			fyne.Do(func() {
				// Обновляем список чатов
				w.refreshChatsFunc(*chat)

				w.window.Close()
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
		searchButton,
		foundUsersLabel,
	)

	bottomContent := container.NewCenter(
		createButton,
	)

	content := container.NewBorder(
		topContent,
		bottomContent,
		nil,
		nil,
		usersList,
	)

	window.SetContent(content)

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
