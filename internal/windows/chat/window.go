package auth

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
)

const (
	title = "KFC Chat"

	width  = 900
	height = 700

	chatsLimit  = 0
	chatsOffset = 0

	chatsListChatLabelText     = "chat"
	chatsLabelText             = "Чаты"
	currentUserSenderLabelText = "Вы"

	newChatButtonText = "Новый чат"
	searchButtonText  = "Поиск"
	logoutButtonText  = "Выйти"

	chatsListChatLabelIndex       = 1
	messagesListBorderIndex       = 0
	messagesListSenderLabelIndex  = 0
	messagesListTimeLabelIndex    = 1
	messagesListMessageLabelIndex = 1

	chatTitleDefaultName = "Чат #%d"

	messagesLimit = 10 // Лимит сообщений за один запрос

	refreshTokensInterval = 5 * time.Minute
	updateChatsInterval   = 5 * time.Second
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases                                        interfaces.UseCases
	logger                                          logging.Logger
	authWindow, createChatWindow, searchUsersWindow interfaces.Window

	ctx        context.Context
	cancelFunc context.CancelFunc

	wg         sync.WaitGroup
	chatsMu    sync.RWMutex
	messagesMu sync.RWMutex

	currentUser     *domains.User
	currentChat     *domains.Chat
	chats           []domains.Chat
	messages        []domains.Message
	hasMoreMessages bool

	chatsList              *widget.List
	messagesList           *widget.List
	messageEntry           *widget.Entry
	loadMoreMessagesButton *widget.Button
	rightPanel             *fyne.Container
}

func New(
	app fyne.App,
	authWindow, createChatWindow, searchUsersWindow interfaces.Window,
	logger logging.Logger,
	useCases interfaces.UseCases,
) *Window {
	return &Window{
		app:               app,
		useCases:          useCases,
		authWindow:        authWindow,
		createChatWindow:  createChatWindow,
		searchUsersWindow: searchUsersWindow,
		logger:            logger,
	}
}

func (w *Window) Build(_ fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	w.ctx, w.cancelFunc = context.WithCancel(context.Background())

	// Устанавливаем обработчик закрытия окна чата
	window.SetCloseIntercept(func() {
		// Закрываем текущее окно
		w.window.Close()

		// Запускаем отмену конгтекста для остановки горутин
		w.cancelFunc()

		// Ожидаем завершения горутин
		w.wg.Wait()

		w.app.Quit()
	})

	var err error

	w.currentUser, err = w.useCases.GetCurrentUser(w.ctx)
	if err != nil {
		logging.LogErrorContext(w.ctx, w.logger, "ошибка получения пользователя", err)

		return
	}

	w.chats, err = w.useCases.GetUserChats(w.ctx, chatsLimit, chatsOffset)
	if err != nil {
		logging.LogErrorContext(w.ctx, w.logger, "ошибка загрузки чатов", err)
	}

	w.buildChatsList()

	w.buildMessagesList()

	leftPanel := w.buildLeftPanel()

	//// Поле ввода
	//messageEntry := widget.NewMultiLineEntry()
	//messageEntry.SetPlaceHolder("Введите сообщение...")
	//messageEntry.OnSubmitted = func(s string) {
	//	w.sendMessage()
	//}
	//
	//sendBtn := widget.NewButtonWithIcon("Отправить", theme.MailSendIcon(), func() {
	//	w.sendMessage()
	//})
	//
	//// Кнопка загрузки истории
	//loadMoreBtn := widget.NewButtonWithIcon("Загрузить историю", theme.ContentAddIcon(), func() {
	//	w.loadMoreMessages()
	//})
	//loadMoreBtn.Hidden = true
	//
	//// Кнопка закрытия чата
	//closeChatBtn := widget.NewButtonWithIcon("Закрыть чат", theme.CancelIcon(), func() {
	//	w.closeChat()
	//})
	//closeChatBtn.Importance = widget.DangerImportance

	//// Верхняя панель для сообщений
	//messagesHeader := container.NewHBox(
	//	widget.NewLabelWithStyle("Сообщения", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	//	layout.NewSpacer(), // Это сдвигает все последующие элементы вправо
	//	loadMoreBtn,
	//	layout.NewSpacer(), // Это сдвигает все последующие элементы вправо
	//	closeChatBtn,
	//)

	//
	//inputArea := container.NewBorder(nil, nil, nil, sendBtn, w.messageEntry)
	//
	//rightPanel := container.NewBorder(
	//	messagesHeader,
	//	inputArea,
	//	nil, nil,
	//	container.NewBorder(
	//		nil, nil, nil, nil,
	//		w.messagesList,
	//	),
	//)
	//rightPanel.Hide() // скрываем, пока не выбран чат
	//
	//w.rightPanel = rightPanel
	//
	//split := container.NewHSplit(leftPanel, rightPanel)
	//split.Offset = 0.25
	//
	//w.window.SetContent(split)
	//
	//// Сохраняем ссылку на кнопку для управления видимостью
	//w.loadMoreBtn = loadMoreBtn
	//w.closeChatBtn = closeChatBtn
	//
	//// Запускаем WebSocket соединение (одно для всех чатов)
	//go func() {
	//	err := connectWebSocket(w.tokens.AccessToken, func(msg *Message) {
	//		w.onNewMessage(msg)
	//	})
	//	if err != nil {
	//		w.showError("Ошибка WebSocket", err)
	//	}
	//}()
	//
	//// Запускаем периодическое обновление списка чатов
	//go w.startChatsRefreshRoutine()

	w.window = window
}

func (w *Window) Show() {
	w.startRefreshTokensGoroutine()
	w.startUpdateChatsGoroutine()
	w.startReadMessagesGoroutine()

	w.window.Show()
}

func (w *Window) Close() {
	w.window.Close()
}

func (w *Window) startRefreshTokensGoroutine() {
	w.wg.Go(func() {
		ticker := time.NewTicker(refreshTokensInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				logging.LogInfo(w.logger, "startTokenRefreshRoutine завершена")

				return
			case <-ticker.C:
				if _, err := w.useCases.RefreshTokens(w.ctx); err != nil {
					logging.LogErrorContext(w.ctx, w.logger, "Не удалось обновить токены пользователя", err)
				}
			}
		}
	})
}

func (w *Window) startUpdateChatsGoroutine() {
	w.wg.Go(func() {
		ticker := time.NewTicker(updateChatsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				logging.LogInfo(w.logger, "startUpdateChatsGoroutine завершена")

				return
			case <-ticker.C:
				if err := w.updateChats(); err != nil {
					logging.LogErrorContext(w.ctx, w.logger, "Не удалось обновить список чатов", err)
				}
			}
		}
	})
}

func (w *Window) updateChats() error {
	chats, err := w.useCases.GetUserChats(w.ctx, chatsLimit, chatsOffset)
	if err != nil {
		return err
	}

	w.chatsMu.Lock()
	defer w.chatsMu.Unlock()

	w.chats = chats

	fyne.Do(func() {
		w.chatsList.Refresh()
	})

	return nil
}

func (w *Window) startReadMessagesGoroutine() {
	w.wg.Go(func() {
		for {
			select {
			case <-w.ctx.Done():
				logging.LogInfo(w.logger, "startReadMessagesGoroutine завершена")

				return
			default:
				message, err := w.useCases.ReadMessage(w.ctx)
				if err != nil {
					// Соккет закрыт, отключаем горутину
					if errors.Is(err, customerrors.ErrWebsocketClosed) {
						logging.LogInfo(w.logger, "startReadMessagesGoroutine завершена из-за закрытия соединения")

						return
					}

					logging.LogErrorContext(w.ctx, w.logger, "Не удалось прочитать сообщение", err)

					break
				}

				w.readMessage(*message)
			}
		}
	})
}

func (w *Window) readMessage(message domains.Message) {
	//отбраать соощение для юзера, который его отправил, не нужно
	if message.Sender.ID == w.currentUser.ID {
		return
	}

	// Проверяем, есть ли чат в списке
	chatExists := false

	w.chatsMu.RLock()

	for _, chat := range w.chats {
		if chat.ID == message.ChatID {
			chatExists = true

			break
		}
	}

	w.chatsMu.RUnlock()

	// Если чата нет в списке - обновляем весь список
	if !chatExists {
		go func() {
			if err := w.updateChats(); err != nil {
				logging.LogErrorContext(w.ctx, w.logger, "не удалось обновить чаты", err)
			}
		}()
	} else {
		// Если чат есть и это НЕ текущий чат - помечаем как непрочитанный
		if w.currentChat == nil || message.ChatID != w.currentChat.ID {
			w.chatsMu.Lock()
			for i := range w.chats {
				if w.chats[i].ID == message.ChatID {
					w.chats[i].IsRead = false

					break
				}
			}
			w.chatsMu.Unlock()

			fyne.Do(func() {
				w.chatsList.Refresh()
			})
		}
	}

	// Если сообщение для текущего чата - добавляем его
	if w.currentChat != nil && message.ChatID == w.currentChat.ID {
		w.messagesMu.Lock()

		w.messages = append(w.messages, message)
		w.messagesMu.Unlock()

		fyne.Do(func() {
			w.messagesList.Refresh()
			w.messagesList.ScrollToBottom()
		})

		// Помечаем чат как прочитанный
		w.markChatAsRead(message.ChatID)
	} else if message.Sender.ID != w.currentUser.ID {
		// Показываем уведомление для сообщений из других чатов
		chatTitle := w.getChatTitle(message.ChatID)

		showNotification(cw.app, "Новое сообщение", fmt.Sprintf("[%s] %s: %s",
			chatTitle, message.Sender.Username, message.Text))
	}
}

func (w *Window) getChatTitle(chatID uint64) string {
	w.chatsMu.RLock()
	defer w.chatsMu.RUnlock()

	for _, chat := range w.chats {
		if chat.ID == chatID {
			if chat.Title != nil && *chat.Title != "" {
				return *chat.Title
			}

			if len(chat.Members) > 0 {
				for _, m := range chat.Members {
					if m.ID != w.currentUser.ID {
						return m.Username
					}
				}
			}
		}
	}

	return fmt.Sprintf(chatTitleDefaultName, chatID)
}

func (w *Window) buildChatsList() {
	w.chatsList = widget.NewList(
		func() int {
			w.chatsMu.RLock()
			defer w.chatsMu.RUnlock()

			return len(w.chats)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.MailComposeIcon()),
				widget.NewLabel(chatsListChatLabelText),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			w.chatsMu.RLock()
			defer w.chatsMu.RUnlock()

			if id >= len(w.chats) {
				return
			}

			chat := w.chats[id]
			container_ := item.(*fyne.Container)
			chatTitleLabel := container_.Objects[chatsListChatLabelIndex].(*widget.Label)

			chatTitle := fmt.Sprintf(chatTitleDefaultName, chat.ID)

			if chat.Title != nil {
				switch {
				case *chat.Title != "":
					chatTitle = *chat.Title
				case len(chat.Members) > 0:
					for _, m := range chat.Members {
						if m.ID != w.currentUser.ID {
							chatTitle = m.Username

							break
						}
					}
				}
			}

			chatTitleLabel.SetText(chatTitle)

			if !chat.IsRead {
				chatTitleLabel.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				chatTitleLabel.TextStyle = fyne.TextStyle{}
			}

			// Принудительный рефреш для отображения нового чата жирным в списке
			chatTitleLabel.Refresh()
		},
	)

	w.chatsList.OnSelected = func(id widget.ListItemID) {
		w.chatsMu.RLock()
		defer w.chatsMu.RUnlock()

		if id >= len(w.chats) {
			return
		}

		chat := w.chats[id]

		w.selectChat(chat)
	}
}

func (w *Window) selectChat(chat domains.Chat) {
	w.currentChat = &chat

	// Помечаем чат как прочитанный
	w.markChatAsRead(chat.ID)

	w.messages = nil
	w.hasMoreMessages = true

	messages, err := w.useCases.GetChatMessages(w.ctx, chat.ID, messagesLimit, 0)
	if err != nil {
		dialog.ShowError(err, w.window)

		return
	}

	// Если сообщений меньше лимита, значит больше сообщений нет и подгружать не надо
	if len(messages) < messagesLimit {
		w.hasMoreMessages = false
	}

	w.messagesMu.Lock()
	slices.Reverse(messages)
	w.messages = messages
	w.messagesMu.Unlock()

	fyne.Do(func() {
		w.rightPanel.Show()

		if w.loadMoreMessagesButton != nil {
			switch w.hasMoreMessages {
			case true:
				w.loadMoreMessagesButton.Enable()
				w.loadMoreMessagesButton.Show()
			case false:
				w.loadMoreMessagesButton.Disable()
				w.loadMoreMessagesButton.Hide()
			}
		}

		if len(w.messages) > 0 {
			w.messagesList.ScrollToBottom()
		}

		w.messagesList.Refresh()
	})
}

func (w *Window) markChatAsRead(id uint64) {
	w.chatsMu.Lock()
	defer w.chatsMu.Unlock()

	for i := range w.chats {
		if w.chats[i].ID == id {
			// TODO нужно будет обращаться к ручке MarkChatRead помимо отметки о прочитанности в UI
			w.chats[i].IsRead = true

			break
		}
	}

	fyne.Do(func() {
		w.chatsList.Refresh()
	})
}

func (w *Window) buildMessagesList() {
	w.messagesList = widget.NewList(
		func() int { return len(w.messages) },
		func() fyne.CanvasObject {
			senderLabel := widget.NewLabelWithStyle(
				"",
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			)

			messageLabel := widget.NewLabel("")
			messageLabel.Wrapping = fyne.TextWrapWord

			timeLabel := widget.NewLabelWithStyle(
				"",
				fyne.TextAlignTrailing,
				fyne.TextStyle{Italic: true},
			)

			return container.NewVBox(
				container.NewBorder(
					nil,
					nil,
					senderLabel,
					timeLabel,
				),
				messageLabel,
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(w.messages) {
				return
			}

			message := w.messages[id]

			container_ := item.(*fyne.Container)
			border := container_.Objects[messagesListBorderIndex].(*fyne.Container)
			senderLabel := border.Objects[messagesListSenderLabelIndex].(*widget.Label)
			timeLabel := border.Objects[messagesListTimeLabelIndex].(*widget.Label)
			messageLabel := container_.Objects[messagesListMessageLabelIndex].(*widget.Label)

			if message.Sender.ID == w.currentUser.ID {
				senderLabel.SetText(currentUserSenderLabelText)
			} else {
				senderLabel.SetText(message.Sender.Username)
			}

			messageLabel.SetText(message.Text)
			timeLabel.SetText(message.CreatedAt.Format(common.DateTimeFormat))
		},
	)
}

func (w *Window) buildLeftPanel() *fyne.Container {
	newChatButton := widget.NewButtonWithIcon(
		newChatButtonText,
		theme.ContentAddIcon(),
		func() {
			w.createChatWindow.Build(nil)
			w.createChatWindow.Show()
		},
	)

	searchButton := widget.NewButtonWithIcon(
		searchButtonText,
		theme.SearchIcon(),
		func() {
			w.searchUsersWindow.Build(nil)
			w.searchUsersWindow.Show()
		},
	)

	logoutButton := widget.NewButtonWithIcon(
		logoutButtonText,
		theme.LogoutIcon(),
		func() {
			w.logout()
		},
	)
	logoutButton.Importance = widget.DangerImportance

	toolbar := container.NewVBox(
		newChatButton,
		searchButton,
	)

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle(
				chatsLabelText,
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			),
			widget.NewSeparator(),
		),
		logoutButton,
		nil,
		nil,
		container.NewBorder(
			toolbar,
			nil,
			nil,
			nil,
			w.chatsList,
		),
	)
}

func (w *Window) logout() {
	w.window.Close()
	if err := w.useCases.Logout(w.ctx); err != nil {
		logging.LogErrorContext(w.ctx, w.logger, "не удалось сделать Logout", err)
	}

	w.cancelFunc()

	w.wg.Wait()

	w.authWindow.Build(nil)
	w.authWindow.Show()
}
