package chat

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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

	chatsListChatTitleLabelText              = "chat"
	chatsListNewChatIndicatorLabelText       = "●"
	chatsLabelText                           = "Чаты"
	currentUserSenderLabelText               = "Вы"
	messagesHeaderLabelText                  = "Сообщения"
	messagesListNewMessageIndicatorLabelText = "●"

	newChatButtonText          = "Новый чат"
	searchButtonText           = "Поиск"
	logoutButtonText           = "Выйти"
	loadMoreMessagesButtonText = "Загрузить историю"
	closeChatButtonText        = "Закрыть чат"
	sendMessageButtonText      = "Отправить"
	changeThemeButtonText      = "Сменить тему"

	messageEntryText = "Введите сообщение..."

	chatsListChatTitleLabelIndex              = 1
	chatListNewChatIndicatorLabelIndex        = 2
	messagesListHeaderIndex                   = 0
	messagesListSenderLabelIndex              = 0
	messagesListHeaderRightPartBoxIndex       = 1
	messagesListTimeLabelIndex                = 0
	messagesListNewMessageIndicatorLabelIndex = 1
	messagesListMessageLabelIndex             = 1

	chatTitleDefaultName = "Чат #%d"

	messagesLimit = 10 // Лимит сообщений за один запрос

	refreshTokensInterval = 1 * time.Minute
	updateChatsInterval   = 5 * time.Second

	panelsSplitOffset = 0.25

	additionalMessageHeight = 30
)

type Window struct {
	app    fyne.App
	window fyne.Window

	useCases                                                            interfaces.UseCases
	logger                                                              logging.Logger
	authWindow, createChatWindow, searchUsersWindow, notificationWindow interfaces.Window

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
	sendMessageButton      *widget.Button
	rightPanel             *fyne.Container

	minMessageSize float32
}

func New(
	app fyne.App,
	authWindow, createChatWindow, searchUsersWindow, notificationWindow interfaces.Window,
	logger logging.Logger,
	useCases interfaces.UseCases,
) *Window {
	return &Window{
		app:                app,
		useCases:           useCases,
		authWindow:         authWindow,
		createChatWindow:   createChatWindow,
		searchUsersWindow:  searchUsersWindow,
		notificationWindow: notificationWindow,
		logger:             logger,
	}
}

func (w *Window) Build(_ fyne.CanvasObject) {
	window := w.app.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	w.ctx, w.cancelFunc = context.WithCancel(context.Background())

	// Устанавливаем обработчик закрытия окна чата
	window.SetCloseIntercept(func() {
		w.Close()

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
	w.rightPanel = w.buildRightPanel()

	split := container.NewHSplit(leftPanel, w.rightPanel)
	split.Offset = panelsSplitOffset

	window.SetContent(split)

	w.window = window
}

func (w *Window) Show() {
	if w.window == nil {
		return
	}

	w.startRefreshTokensGoroutine()
	w.startUpdateChatsGoroutine()
	w.startReadMessagesGoroutine()

	w.window.Show()
}

func (w *Window) Close() {
	if w.window == nil {
		return
	}

	// Закрываем текущее окно
	w.window.Close()

	// Запускаем отмену конгтекста для остановки горутин
	w.cancelFunc()

	// Ожидаем завершения горутин
	w.wg.Wait()

	w.window = nil
}

func (w *Window) RefreshChats(chat domains.Chat) {
	w.chatsMu.Lock()
	defer w.chatsMu.Unlock()

	// Проверяем, есть ли уже такой чат в списке
	for i := range w.chats {
		if w.chats[i].ID == chat.ID {
			return
		}
	}

	// Добавляем новый чат в начало списка
	w.chats = append([]domains.Chat{chat}, w.chats...)

	fyne.Do(func() {
		w.chatsList.Refresh()
	})
}

func (w *Window) startRefreshTokensGoroutine() {
	// TODO fyne-cross не умеет в wg.Go(). Перейти в дальнейшем
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(refreshTokensInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				logging.LogInfo(w.logger, "startTokenRefreshRoutine завершена")

				return
			case <-ticker.C:
				if _, err := w.useCases.RefreshTokens(w.ctx); err != nil {
					logging.LogErrorContext(
						w.ctx,
						w.logger,
						"Не удалось обновить токены пользователя",
						err,
					)
				}
			}
		}
	}()
}

func (w *Window) startUpdateChatsGoroutine() {
	// TODO fyne-cross не умеет в wg.Go(). Перейти в дальнейшем
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(updateChatsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				logging.LogInfo(w.logger, "startUpdateChatsGoroutine завершена")

				return
			case <-ticker.C:
				if err := w.updateChats(); err != nil {
					logging.LogErrorContext(
						w.ctx,
						w.logger,
						"Не удалось обновить список чатов",
						err,
					)
				}
			}
		}
	}()
}

func (w *Window) updateChats() error {
	chats, err := w.useCases.GetUserChats(w.ctx, chatsLimit, chatsOffset)
	if err != nil {
		return err
	}

	w.chatsMu.Lock()
	w.chats = chats
	w.chatsMu.Unlock()

	// Если текущий открытый чат был удален, закрываем вкладку
	if w.currentChat != nil {
		chatDeleted := true

		for i := range chats {
			if w.currentChat.ID == chats[i].ID {
				chatDeleted = false

				break
			}
		}

		if chatDeleted {
			fyne.Do(func() {
				w.closeChat()
			})
		}
	}

	fyne.Do(func() {
		w.chatsList.Refresh()
	})

	return nil
}

func (w *Window) startReadMessagesGoroutine() {
	// TODO fyne-cross не умеет в wg.Go(). Перейти в дальнейшем
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

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
						logging.LogInfo(
							w.logger,
							"startReadMessagesGoroutine завершена из-за закрытия соединения",
						)

						return
					}

					logging.LogErrorContext(w.ctx, w.logger, "Не удалось прочитать сообщение", err)

					break
				}

				w.readMessage(*message)
			}
		}
	}()
}

func (w *Window) readMessage(message domains.Message) {
	// отбраать соощение для юзера, который его отправил, не нужно
	if message.Sender.ID == w.currentUser.ID {
		return
	}

	w.markChatUnreadAfterReceivingMessage(message)

	w.processMessage(message)
}

func (w *Window) markChatUnreadAfterReceivingMessage(message domains.Message) {
	// Если чата нет в списке - обновляем весь список
	if !w.chatExists(message) {
		if err := w.updateChats(); err != nil {
			logging.LogErrorContext(w.ctx, w.logger, "не удалось обновить чаты", err)
		}

		return
	}

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

func (w *Window) chatExists(message domains.Message) bool {
	w.chatsMu.RLock()
	defer w.chatsMu.RUnlock()

	for i := range w.chats {
		if w.chats[i].ID == message.ChatID {
			return true
		}
	}

	return false
}

func (w *Window) processMessage(message domains.Message) {
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

		return
	}

	if message.Sender.ID != w.currentUser.ID {
		// Показываем уведомление для сообщений из других чатов
		chatTitle := w.getChatTitle(message.ChatID)

		fyne.Do(func() {
			w.notificationWindow.Build(
				widget.NewLabel(
					fmt.Sprintf(
						"[%s] %s: %s",
						chatTitle,
						message.Sender.Username,
						message.Text,
					),
				),
			)
			w.notificationWindow.Show()
		})
	}
}

func (w *Window) getChatTitle(chatID uint64) string {
	w.chatsMu.RLock()
	defer w.chatsMu.RUnlock()

	for _, chat := range w.chats { //nolint:gocritic // простота чтения важнее
		if chat.ID != chatID {
			continue
		}

		if chat.Title != nil && *chat.Title != "" {
			return *chat.Title
		}

		if chat.Type == domains.ChatTypePrivate {
			for _, m := range chat.Members {
				if m.ID != w.currentUser.ID {
					return m.Username
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
			// Индикатор нового сообщения
			newChatIndicatorLabel := widget.NewLabel(chatsListNewChatIndicatorLabelText)
			newChatIndicatorLabel.Hide()
			newChatIndicatorLabel.TextStyle = fyne.TextStyle{Bold: true}

			return container.NewHBox(
				widget.NewIcon(
					theme.AccountIcon(),
				), // TODO если чат груповой, то иконку ставим theme.FolderIcon()
				widget.NewLabel(chatsListChatTitleLabelText),
				newChatIndicatorLabel,
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
			chatTitleLabel := container_.Objects[chatsListChatTitleLabelIndex].(*widget.Label)
			newChatIndicatorLabel := container_.Objects[chatListNewChatIndicatorLabelIndex].(*widget.Label)

			chatTitle := fmt.Sprintf(chatTitleDefaultName, chat.ID)

			switch {
			case chat.Title != nil && *chat.Title != "":
				chatTitle = *chat.Title
			case chat.Type == domains.ChatTypePrivate:
				for _, m := range chat.Members {
					if m.ID != w.currentUser.ID {
						chatTitle = m.Username

						break
					}
				}
			}

			chatTitleLabel.SetText(chatTitle)

			if !chat.IsRead {
				chatTitleLabel.TextStyle = fyne.TextStyle{Bold: true}

				newChatIndicatorLabel.Show()
			} else {
				chatTitleLabel.TextStyle = fyne.TextStyle{}

				newChatIndicatorLabel.Hide()
			}

			// Принудительный рефреш для отображения нового чата жирным в списке
			chatTitleLabel.Refresh()
		},
	)

	w.chatsList.OnSelected = func(id widget.ListItemID) {
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
	w.minMessageSize = 0

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

	w.rightPanel.Show()

	switch w.hasMoreMessages {
	case true:
		w.loadMoreMessagesButton.Enable()
		w.loadMoreMessagesButton.Show()
	case false:
		w.loadMoreMessagesButton.Disable()
		w.loadMoreMessagesButton.Hide()
	}

	w.messagesList.Refresh()

	if len(w.messages) > 0 {
		w.messagesList.ScrollToBottom()
	}
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
		func() int {
			w.messagesMu.RLock()
			defer w.messagesMu.RUnlock()

			return len(w.messages)
		},
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

			// Индикатор нового сообщения
			newMessageIndicatorLabel := widget.NewLabel(messagesListNewMessageIndicatorLabelText)
			newMessageIndicatorLabel.Hide()
			newMessageIndicatorLabel.TextStyle = fyne.TextStyle{Bold: true}

			header := container.NewBorder(
				nil,
				nil,
				senderLabel,
				container.NewHBox(
					timeLabel,
					newMessageIndicatorLabel,
				), // Справа время + индикатор
			)

			return container.NewVBox(
				header,
				messageLabel,
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// Используем w.messagesMu.Unlock() без defer из-за w.messagesList.SetItemHeight() в конце метода
			w.messagesMu.Lock()

			if id >= len(w.messages) {
				w.messagesMu.Unlock()

				return
			}

			message := w.messages[id]

			w.messagesMu.Unlock()

			container_ := item.(*fyne.Container)
			headerBorder := container_.Objects[messagesListHeaderIndex].(*fyne.Container)
			senderLabel := headerBorder.Objects[messagesListSenderLabelIndex].(*widget.Label)
			headerRightPartBox := headerBorder.Objects[messagesListHeaderRightPartBoxIndex].(*fyne.Container)
			timeLabel := headerRightPartBox.Objects[messagesListTimeLabelIndex].(*widget.Label)
			newMessageIndicatorLabel := headerRightPartBox.Objects[messagesListNewMessageIndicatorLabelIndex].(*widget.Label)
			messageLabel := container_.Objects[messagesListMessageLabelIndex].(*widget.Label)

			if message.Sender.ID == w.currentUser.ID {
				senderLabel.SetText(currentUserSenderLabelText)
			} else {
				senderLabel.SetText(message.Sender.Username)
			}

			timeLabel.SetText(message.CreatedAt.Format(common.DateTimeFormat))

			messageLabel.SetText(message.Text)

			if !message.IsRead {
				senderLabel.TextStyle = fyne.TextStyle{Bold: true}
				timeLabel.TextStyle = fyne.TextStyle{Bold: true}
				messageLabel.TextStyle = fyne.TextStyle{Bold: true}
				newMessageIndicatorLabel.TextStyle = fyne.TextStyle{Bold: true}

				senderLabel.Refresh()
				timeLabel.Refresh()
				messageLabel.Refresh()
				newMessageIndicatorLabel.Refresh()

				newMessageIndicatorLabel.Show()
			} else {
				senderLabel.TextStyle = fyne.TextStyle{Bold: false}
				timeLabel.TextStyle = fyne.TextStyle{Bold: false}
				messageLabel.TextStyle = fyne.TextStyle{Bold: false}
				newMessageIndicatorLabel.TextStyle = fyne.TextStyle{Bold: false}

				senderLabel.Refresh()
				timeLabel.Refresh()
				messageLabel.Refresh()
				newMessageIndicatorLabel.Refresh()

				newMessageIndicatorLabel.Hide()
			}

			// Динамически подгоняем высоту итемов, чтобы все помещалось
			if messageLabel.MinSize().Height > w.minMessageSize {
				w.minMessageSize = messageLabel.MinSize().Height
			}

			w.messagesList.SetItemHeight(id, w.minMessageSize+additionalMessageHeight)
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

	bottom := container.NewVBox(
		widget.NewButtonWithIcon(
			changeThemeButtonText,
			theme.ColorPaletteIcon(),
			func() {
				var (
					appTheme         fyne.Theme
					newSettingsTheme domains.ThemeType
				)

				currentTheme := w.useCases.GetTheme(w.ctx)
				switch currentTheme {
				case domains.ThemeLight:
					appTheme = theme.DarkTheme()
					newSettingsTheme = domains.ThemeDark
				case domains.ThemeDark:
					appTheme = theme.LightTheme()
					newSettingsTheme = domains.ThemeLight
				}

				if err := w.useCases.SetTheme(w.ctx, newSettingsTheme); err != nil {
					dialog.ShowError(err, w.window)

					return
				}

				w.app.Settings().SetTheme(appTheme)
			},
		),
		logoutButton,
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
		bottom,
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
	if err := w.useCases.Logout(w.ctx); err != nil {
		logging.LogErrorContext(w.ctx, w.logger, "не удалось сделать Logout", err)
	}

	w.Close()

	w.authWindow.Build(nil)
	w.authWindow.Show()
}

func (w *Window) buildRightPanel() *fyne.Container {
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder(messageEntryText)
	messageEntry.OnSubmitted = func(_ string) {
		w.sendMessage()
	}
	w.messageEntry = messageEntry

	sendMessageButton := widget.NewButtonWithIcon(
		sendMessageButtonText,
		theme.MailSendIcon(),
		func() {
			w.sendMessage()
		},
	)
	sendMessageButton.Importance = widget.SuccessImportance
	w.sendMessageButton = sendMessageButton

	// Кнопка загрузки истории
	loadMoreMessagesButton := widget.NewButtonWithIcon(
		loadMoreMessagesButtonText,
		theme.ContentAddIcon(),
		func() {
			w.loadMoreMessages()
		},
	)
	loadMoreMessagesButton.Hidden = true
	w.loadMoreMessagesButton = loadMoreMessagesButton

	// Кнопка закрытия чата
	closeChatButton := widget.NewButtonWithIcon(
		closeChatButtonText,
		theme.CancelIcon(),
		func() {
			w.closeChat()
		},
	)
	closeChatButton.Importance = widget.DangerImportance

	// Верхняя панель для сообщений
	messagesHeader := container.NewHBox(
		widget.NewLabelWithStyle(
			messagesHeaderLabelText,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		layout.NewSpacer(), // Это сдвигает все последующие элементы вправо
		loadMoreMessagesButton,
		layout.NewSpacer(), // Это сдвигает все последующие элементы вправо
		closeChatButton,
	)

	inputArea := container.NewBorder(
		nil,
		nil,
		nil,
		sendMessageButton,
		w.messageEntry,
	)

	rightPanel := container.NewBorder(
		messagesHeader,
		inputArea,
		nil,
		nil,
		w.messagesList,
	)
	rightPanel.Hide() // скрываем, пока не выбран чат

	return rightPanel
}

func (w *Window) sendMessage() {
	w.sendMessageButton.Disable()
	defer w.sendMessageButton.Enable()

	text := strings.TrimSpace(w.messageEntry.Text)
	if text == "" || w.currentChat == nil {
		return
	}

	message := domains.Message{
		ChatID:    w.currentChat.ID,
		Sender:    *w.currentUser,
		Text:      text,
		CreatedAt: time.Now().In(common.Timezone),
		UpdatedAt: time.Now().In(common.Timezone),
		IsRead:    true, // Сообщение прочитано для отправителя
	}

	if err := w.useCases.SendMessage(w.ctx, message); err != nil {
		dialog.ShowError(err, w.window)

		return
	}

	// Обновляем список сообщбений только после успешной отправки на сервер
	w.messagesMu.Lock()

	// Все сообщения в списке помечаем прочитанными, когда пользователь отправляет в чат новое сообщение
	for i := range w.messages {
		w.messages[i].IsRead = true
	}

	w.messages = append(w.messages, message)
	w.messagesMu.Unlock()

	// Стираем отправленное сообщение из поля ввода
	w.messageEntry.SetText("")
	w.messagesList.Refresh()
	w.messagesList.ScrollToBottom()
}

func (w *Window) loadMoreMessages() {
	if !w.hasMoreMessages || w.currentChat == nil {
		return
	}

	w.loadMoreMessagesButton.Disable()
	w.loadMoreMessagesButton.Hide()

	offset := len(w.messages)

	messages, err := w.useCases.GetChatMessages(w.ctx, w.currentChat.ID, messagesLimit, offset)
	if err != nil {
		dialog.ShowError(err, w.window)
		w.loadMoreMessagesButton.Enable()
		w.loadMoreMessagesButton.Show()

		return
	}

	// Больше нет сообщений в чате, скрываем кнопку подгрузки сообщений
	if len(messages) == 0 {
		w.hasMoreMessages = false

		return
	}

	slices.Reverse(messages)

	// Добавляем старые сообщения в начало
	w.messagesMu.Lock()
	w.messages = append(messages, w.messages...)
	w.messagesMu.Unlock()

	w.messagesList.Refresh()

	// Скроллим к первым новым сообщениям (которые были до загрузки)
	w.messagesList.ScrollTo(len(messages))

	// Если загружено меньше лимита, значит больше нет сообщений
	if len(messages) < messagesLimit {
		w.hasMoreMessages = false

		return
	}

	w.loadMoreMessagesButton.Enable()
	w.loadMoreMessagesButton.Show()
}

func (w *Window) closeChat() {
	if w.currentChat == nil {
		return
	}

	w.rightPanel.Hide()
	w.currentChat = nil
	w.messages = nil
	w.hasMoreMessages = true
	w.minMessageSize = 0

	// Очищаем поле ввода
	w.messageEntry.SetText("")

	// Снимаем выделение
	w.chatsList.UnselectAll()
}
