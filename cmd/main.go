package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/gorilla/websocket"
)

const (
	tokenFile       = "tokens.json"
	apiBase         = "http://localhost:8080"
	wsBase          = "ws://localhost:8080"
	messagesLimit   = 10 // Лимит сообщений за один запрос
	refreshInterval = 5 * time.Minute
)

// --- Структуры для токенов ---

type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// --- Структуры данных из Swagger ---

type User struct {
	ID             uint64    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	EmailConfirmed bool      `json:"emailConfirmed"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Message struct {
	ID        uint64    `json:"id"`
	ChatID    uint64    `json:"chatId"`
	Sender    User      `json:"sender"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type OutgoingMessage struct {
	ChatID uint64 `json:"chatId"`
	Sender User   `json:"sender"`
	Text   string `json:"text"`
}

type Chat struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"`
	Members     []User    `json:"members,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	IsRead      bool      `json:"isRead"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type MemberInput struct {
	ID uint64 `json:"id"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// --- Глобальные переменные ---
// TODO подумать о переводе всех мьютексов и каналов на context.WithCancel.
var (
	httpClient    = &http.Client{Timeout: 10 * time.Second}
	currentUser   *User
	currentTokens *TokenPair
	wsConn        *websocket.Conn
	wsMutex       sync.Mutex
	messagesMu    sync.RWMutex
	refreshTicker *time.Ticker
	refreshDone   chan struct{}
	appShutdown   = make(chan struct{})
	wg            sync.WaitGroup
	shutdownOnce  sync.Once  // Гарантирует, что канал закроется только один раз
	refreshMu     sync.Mutex // Мьютекс для защиты доступа к refreshDone
)

// --- Основная функция ---

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DefaultTheme())

	go authenticateAndRun(myApp)

	myApp.Run()
}

// --- Аутентификация ---

func authenticateAndRun(myApp fyne.App) {
	tokens, err := readTokensFromFile()

	if err == nil && tokens != nil && tokens.AccessToken != "" && tokens.RefreshToken != "" {
		currentTokens = tokens

		user, err := getCurrentUser(tokens.AccessToken)
		if err == nil && user != nil {
			currentUser = user

			startTokenRefreshRoutine()

			fyne.Do(func() {
				showMainChatWindow(myApp, tokens)
			})

			return
		}

		newTokens, refreshErr := refreshTokens(tokens.RefreshToken)
		if refreshErr == nil && newTokens != nil {
			saveTokensToFile(newTokens)
			currentTokens = newTokens

			user, err := getCurrentUser(newTokens.AccessToken)
			if err == nil && user != nil {
				currentUser = user

				startTokenRefreshRoutine()

				fyne.Do(func() {
					showMainChatWindow(myApp, newTokens)
				})

				return
			}
		}
	}

	fyne.Do(func() {
		showAuthWindow(myApp)
	})
}

// --- Работа с токенами ---

func readTokensFromFile() (*TokenPair, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var tokens TokenPair
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

func saveTokensToFile(tokens *TokenPair) error {
	tokens.ExpiresAt = time.Now().Add(4*time.Minute + 30*time.Second)

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0o600)
}

func startTokenRefreshRoutine() {
	refreshTicker = time.NewTicker(refreshInterval)
	refreshDone = make(chan struct{}) // Используем struct{}

	wg.Go(func() {
		for {
			select {
			case <-appShutdown:
				fmt.Println("Token refresh: приложение завершается")

				return
			case <-refreshDone:
				fmt.Println("Token refresh: получен сигнал завершения")

				return
			case <-refreshTicker.C:
				if isShuttingDown() {
					return
				}

				refreshTokenRoutine()
			}
		}
	})
}

func refreshTokenRoutine() {
	if currentTokens == nil || currentTokens.RefreshToken == "" {
		return
	}

	newTokens, err := refreshTokens(currentTokens.RefreshToken)
	if err != nil {
		fmt.Printf("Ошибка обновления токенов: %v\n", err)

		return
	}

	currentTokens = newTokens
	saveTokensToFile(newTokens)

	fmt.Println("Токены успешно обновлены")
}

// --- HTTP запросы к API ---

func getCurrentUser(accessToken string) (*User, error) {
	req, _ := http.NewRequest(http.MethodGet, apiBase+"/users/me", http.NoBody)
	req.Header.Set("Cookie", "accessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func refreshTokens(refreshToken string) (*TokenPair, error) {
	req, _ := http.NewRequest(http.MethodPut, apiBase+"/sessions", http.NoBody)
	req.Header.Set("Cookie", "refreshToken="+refreshToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("refresh failed: %s", resp.Status)
	}

	tokens := &TokenPair{}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "accessToken" {
			tokens.AccessToken = cookie.Value
		}

		if cookie.Name == "refreshToken" {
			tokens.RefreshToken = cookie.Value
		}
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return nil, errors.New("не удалось получить токены из cookie")
	}

	return tokens, nil
}

func login(email, password string) (*TokenPair, error) {
	input := LoginInput{
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("login failed (%s): %s", resp.Status, string(errorBody))
	}

	tokens := &TokenPair{}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "accessToken" {
			tokens.AccessToken = cookie.Value
		}

		if cookie.Name == "refreshToken" {
			tokens.RefreshToken = cookie.Value
		}
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return nil, errors.New("не удалось получить токены из cookie")
	}

	return tokens, nil
}

func logout(accessToken string) error {
	req, _ := http.NewRequest(http.MethodDelete, apiBase+"/sessions", http.NoBody)
	req.Header.Set("Cookie", "accessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout failed: %s", resp.Status)
	}

	return nil
}

func getUserChats(accessToken string, limit, offset int) ([]Chat, error) {
	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/chats?limit=%d&offset=%d", apiBase, limit, offset),
		http.NoBody,
	)
	req.Header.Set("Cookie", "accessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var chats []Chat
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func getChatMessages(accessToken string, chatID uint64, limit, offset int) ([]Message, error) {
	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/chats/%d/messages?limit=%d&offset=%d", apiBase, chatID, limit, offset),
		http.NoBody,
	)
	req.Header.Set("Cookie", "accessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var messages []Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func createChat(accessToken string, chatType string, memberIDs []uint64) (*Chat, error) {
	members := make([]MemberInput, len(memberIDs))
	for i, id := range memberIDs {
		members[i] = MemberInput{ID: id}
	}

	bodyMap := map[string]any{
		"type":    chatType,
		"members": members,
	}
	body, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/chats", bytes.NewReader(body))
	req.Header.Set("Cookie", "accessToken="+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(errorBody))
	}

	var chat Chat
	if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

func searchUsers(accessToken, query string, limit, offset int) ([]User, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?username=%s&limit=%d&offset=%d",
		apiBase, url.QueryEscape(query), limit, offset), http.NoBody)
	req.Header.Set("Cookie", "accessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

// --- WebSocket подключение ---.
func connectWebSocket(accessToken string, onMessage func(*Message)) error {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	// Если приложение завершается, не подключаемся
	if isShuttingDown() {
		return errors.New("приложение завершается")
	}

	if wsConn != nil {
		return nil
	}

	header := http.Header{}
	header.Add("Cookie", "accessToken="+accessToken)

	conn, _, err := websocket.DefaultDialer.Dial(wsBase+"/ws", header)
	if err != nil {
		return err
	}

	wsConn = conn

	wg.Go(func() {
		// Канал для сигнала завершения чтения
		readDone := make(chan struct{})

		// Горутина для чтения сообщений
		go func() {
			defer close(readDone)

			for {
				// Проверяем завершение перед каждым чтением
				select {
				case <-appShutdown:
					fmt.Println("WebSocket reader: получен сигнал завершения")

					return
				default:
				}

				var msg Message

				err := conn.ReadJSON(&msg)
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
						errors.Is(err, net.ErrClosed) {
						return
					}

					fmt.Printf("WebSocket ошибка: %v\n", err)

					return
				}

				onMessage(&msg)
			}
		}()

		// Ждем либо завершения чтения, либо сигнала остановки
		select {
		case <-readDone:
			fmt.Println("WebSocket: чтение завершено")
		case <-appShutdown:
			fmt.Println("WebSocket: принудительное завершение")
		}

		// Закрываем соединение
		wsMutex.Lock()
		if wsConn != nil {
			wsConn.Close()
			wsConn = nil
		}
		wsMutex.Unlock()
	})

	return nil
}

func sendWSMessage(chatID uint64, text string) error {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	if wsConn == nil {
		return errors.New("websocket not connected")
	}

	outgoingMsg := OutgoingMessage{
		ChatID: chatID,
		Sender: *currentUser,
		Text:   text,
	}

	return wsConn.WriteJSON(outgoingMsg)
}

func disconnectWebSocket() {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	if wsConn != nil {
		wsConn.Close()
		wsConn = nil
	}
}

// --- Основное окно чата ---

type ChatWindow struct {
	app          fyne.App
	tokens       *TokenPair
	window       fyne.Window
	chatsList    *widget.List
	messagesList *widget.List
	messageEntry *widget.Entry
	currentChat  *Chat
	chats        []Chat
	messages     []Message
	loadMoreBtn  *widget.Button // Добавлено
	closeChatBtn *widget.Button
	loadingMore  bool
	hasMore      bool
	chatsMu      sync.RWMutex // Мьютекс для защиты списка чатов
	rightPanel   *fyne.Container
}

func showMainChatWindow(myApp fyne.App, tokens *TokenPair) {
	cw := &ChatWindow{
		app:     myApp,
		tokens:  tokens,
		window:  myApp.NewWindow("KFC Chat - " + currentUser.Username),
		hasMore: true,
	}

	cw.window.Resize(fyne.NewSize(900, 700))

	cw.window.SetCloseIntercept(func() {
		shutdown(myApp)
	})

	// Список чатов
	cw.chatsList = widget.NewList(
		func() int {
			cw.chatsMu.RLock()
			defer cw.chatsMu.RUnlock()

			return len(cw.chats)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.MailComposeIcon()),
				widget.NewLabel("Chat"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			cw.chatsMu.RLock()
			defer cw.chatsMu.RUnlock()

			if id >= len(cw.chats) {
				return
			}

			chat := cw.chats[id]
			container := item.(*fyne.Container)
			label := container.Objects[1].(*widget.Label)

			title := chat.Title
			if title == "" && len(chat.Members) > 0 {
				for _, m := range chat.Members {
					if m.ID != currentUser.ID {
						title = m.Username

						break
					}
				}
			}

			if title == "" {
				title = fmt.Sprintf("Чат #%d", chat.ID)
			}

			label.SetText(title)

			if !chat.IsRead {
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	cw.chatsList.OnSelected = func(id widget.ListItemID) {
		cw.chatsMu.RLock()

		if id >= len(cw.chats) {
			cw.chatsMu.RUnlock()

			return
		}

		chat := cw.chats[id]
		cw.chatsMu.RUnlock()
		cw.selectChat(&chat)
	}

	// Список сообщений
	cw.messagesList = widget.NewList(
		func() int { return len(cw.messages) },
		func() fyne.CanvasObject {
			sender := widget.NewLabelWithStyle(
				"",
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			)
			message := widget.NewLabel("")
			message.Wrapping = fyne.TextWrapWord
			time_ := widget.NewLabelWithStyle(
				"",
				fyne.TextAlignTrailing,
				fyne.TextStyle{Italic: true},
			)

			return container.NewVBox(
				container.NewBorder(nil, nil, sender, time_),
				message,
				widget.NewSeparator(),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(cw.messages) {
				return
			}

			msg := cw.messages[id]

			container_ := item.(*fyne.Container)
			border := container_.Objects[0].(*fyne.Container)
			sender := border.Objects[0].(*widget.Label)
			time_ := border.Objects[1].(*widget.Label)
			message := container_.Objects[1].(*widget.Label)

			if msg.Sender.ID == currentUser.ID {
				sender.SetText("Вы")
			} else {
				sender.SetText(msg.Sender.Username)
			}

			message.SetText(msg.Text)
			time_.SetText(msg.CreatedAt.Format(common.DateTimeFormat))
		},
	)

	// Поле ввода
	cw.messageEntry = widget.NewMultiLineEntry()
	cw.messageEntry.SetPlaceHolder("Введите сообщение...")
	cw.messageEntry.OnSubmitted = func(s string) {
		cw.sendMessage()
	}

	sendBtn := widget.NewButtonWithIcon("Отправить", theme.MailSendIcon(), func() {
		cw.sendMessage()
	})

	// Кнопка загрузки истории
	loadMoreBtn := widget.NewButtonWithIcon("Загрузить историю", theme.ContentAddIcon(), func() {
		cw.loadMoreMessages()
	})
	loadMoreBtn.Hidden = true

	// Кнопка закрытия чата
	closeChatBtn := widget.NewButtonWithIcon("Закрыть чат", theme.CancelIcon(), func() {
		cw.closeChat()
	})
	closeChatBtn.Hidden = true
	closeChatBtn.Importance = widget.DangerImportance

	// Кнопки управления
	newChatBtn := widget.NewButtonWithIcon("Новый чат", theme.ContentAddIcon(), func() {
		showCreateChatDialog(cw)
	})

	searchBtn := widget.NewButtonWithIcon("Поиск", theme.SearchIcon(), func() {
		showSearchUsersDialog(cw)
	})

	logoutBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		cw.logout()
	})
	logoutBtn.Importance = widget.DangerImportance

	// Верхняя панель для сообщений
	messagesHeader := container.NewHBox(
		widget.NewLabelWithStyle("Сообщения", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(), // Это pushes все последующие элементы вправо
		loadMoreBtn,
		layout.NewSpacer(), // Это pushes все последующие элементы вправо
		closeChatBtn,
	)

	// Левая панель с чатами и кнопками
	leftToolbar := container.NewVBox(
		newChatBtn,
		searchBtn,
	)

	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Чаты", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
		),
		logoutBtn,
		nil, nil,
		container.NewBorder(
			leftToolbar,
			nil, nil, nil,
			cw.chatsList,
		),
	)

	inputArea := container.NewBorder(nil, nil, nil, sendBtn, cw.messageEntry)

	rightPanel := container.NewBorder(
		messagesHeader,
		inputArea,
		nil, nil,
		container.NewBorder(
			nil, nil, nil, nil,
			cw.messagesList,
		),
	)
	rightPanel.Hide() // скрываем, пока не выбран чат

	cw.rightPanel = rightPanel

	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.25

	cw.window.SetContent(split)

	// Сохраняем ссылку на кнопку для управления видимостью
	cw.loadMoreBtn = loadMoreBtn
	cw.closeChatBtn = closeChatBtn

	cw.loadChats()

	// Запускаем WebSocket соединение (одно для всех чатов)
	go func() {
		err := connectWebSocket(cw.tokens.AccessToken, func(msg *Message) {
			cw.onNewMessage(msg)
		})
		if err != nil {
			cw.showError("Ошибка WebSocket", err)
		}
	}()

	// Запускаем периодическое обновление списка чатов
	go cw.startChatsRefreshRoutine()

	cw.window.Show()
}

func (cw *ChatWindow) loadChats() {
	// TODO при переносе в репозиторий сделать передачу лимитов и оффсета
	chats, err := getUserChats(cw.tokens.AccessToken, 0, 0)
	if err != nil {
		cw.showError("Ошибка загрузки чатов", err)

		return
	}

	cw.chats = chats

	fyne.Do(func() {
		cw.chatsList.Refresh()
	})
}

// Периодическое обновление списка чатов (каждые 5 секунд).
func (cw *ChatWindow) startChatsRefreshRoutine() {
	wg.Go(func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-appShutdown:
				fmt.Println("Chats refresh: получен сигнал завершения")

				return
			case <-ticker.C:
				if cw.currentChat == nil {
					continue
				}

				cw.refreshChats()
			}
		}
	})
}

// Обновление списка чатов.
func (cw *ChatWindow) refreshChats() {
	if isShuttingDown() {
		return
	}

	// TODO при переносе в репозиторий сделать передачу лимитов и оффсета
	chats, err := getUserChats(cw.tokens.AccessToken, 0, 0)
	if err != nil {
		fmt.Printf("Ошибка обновления списка чатов: %v\n", err)

		return
	}

	cw.chatsMu.Lock()
	defer cw.chatsMu.Unlock()

	// Сохраняем информацию о прочитанности
	oldChats := make(map[uint64]bool)
	for _, chat := range cw.chats {
		oldChats[chat.ID] = chat.IsRead
	}

	// Обновляем список, сохраняя статус прочитанности
	for i, newChat := range chats {
		if isRead, exists := oldChats[newChat.ID]; exists {
			chats[i].IsRead = isRead
		}
	}

	cw.chats = chats

	fyne.Do(func() {
		cw.chatsList.Refresh()
	})
}

// Обновление списка после создания нового чата.
func (cw *ChatWindow) refreshChatsAfterNewChat(newChat *Chat) {
	cw.chatsMu.Lock()
	defer cw.chatsMu.Unlock()

	// Проверяем, есть ли уже такой чат
	for _, chat := range cw.chats {
		if chat.ID == newChat.ID {
			return // Чат уже есть в списке
		}
	}

	// Добавляем новый чат в начало списка
	cw.chats = append([]Chat{*newChat}, cw.chats...)

	fyne.Do(func() {
		cw.chatsList.Refresh()
	})
}

func (cw *ChatWindow) loadMoreMessages() {
	if cw.loadingMore || !cw.hasMore || cw.currentChat == nil {
		return
	}

	cw.loadingMore = true

	if cw.loadMoreBtn != nil {
		cw.loadMoreBtn.Disable()
	}

	offset := len(cw.messages)

	go func() {
		messages, err := getChatMessages(
			cw.tokens.AccessToken,
			cw.currentChat.ID,
			messagesLimit,
			offset,
		)
		if err != nil {
			fyne.Do(func() {
				cw.loadingMore = false
				if cw.loadMoreBtn != nil {
					cw.loadMoreBtn.Enable()
				}
			})

			return
		}

		slices.Reverse(messages)

		fyne.Do(func() {
			if len(messages) == 0 {
				cw.hasMore = false

				if cw.loadMoreBtn != nil {
					cw.loadMoreBtn.Hide()
				}
			} else {
				// Добавляем старые сообщения в начало
				messagesMu.Lock()

				cw.messages = append(messages, cw.messages...)
				messagesMu.Unlock()

				cw.messagesList.Refresh()

				// Скроллим к первым новым сообщениям (которые были до загрузки)
				if len(messages) > 0 {
					cw.messagesList.ScrollTo(len(messages))
				}

				// Если загружено меньше лимита, значит больше нет сообщений
				if len(messages) < messagesLimit {
					cw.hasMore = false
					if cw.loadMoreBtn != nil {
						cw.loadMoreBtn.Hide()
					}
				}
			}

			cw.loadingMore = false
			if cw.loadMoreBtn != nil && cw.hasMore {
				cw.loadMoreBtn.Enable()
			}
		})
	}()
}

func (cw *ChatWindow) selectChat(chat *Chat) {
	cw.currentChat = chat

	// Помечаем чат как прочитанный
	cw.markChatAsRead(chat.ID)

	cw.rightPanel.Show()
	cw.messages = nil
	cw.hasMore = true
	cw.loadingMore = false
	cw.messagesList.Refresh()

	// Показываем кнопки
	if cw.loadMoreBtn != nil {
		cw.loadMoreBtn.Enable()
		cw.loadMoreBtn.Show()
	}

	if cw.closeChatBtn != nil {
		cw.closeChatBtn.Show()
	}

	// Загружаем сообщения
	go cw.loadMessages(chat.ID, 0)
}

// --- Обновленный closeChat (без currentChatID) ---.
func (cw *ChatWindow) closeChat() {
	if cw.currentChat == nil {
		return
	}

	// Просто очищаем текущий чат, НЕ закрываем WebSocket
	cw.rightPanel.Hide()
	cw.currentChat = nil
	cw.messages = nil
	cw.hasMore = true
	cw.messagesList.Refresh()

	// Скрываем кнопки
	if cw.loadMoreBtn != nil {
		cw.loadMoreBtn.Enable()
		cw.loadMoreBtn.Hide()
	}

	if cw.closeChatBtn != nil {
		cw.closeChatBtn.Hide()
	}

	// Очищаем поле ввода
	cw.messageEntry.SetText("")

	// Снимаем выделение
	cw.chatsList.UnselectAll()
}

func (cw *ChatWindow) markChatAsRead(chatID uint64) {
	cw.chatsMu.Lock()
	defer cw.chatsMu.Unlock()

	for i, chat := range cw.chats {
		if chat.ID == chatID {
			// TODO нужно будет обращаться к ручке MarkChatRead помимо отметки о прочитанности в UI
			cw.chats[i].IsRead = true

			break
		}
	}

	fyne.Do(func() {
		cw.chatsList.Refresh()
	})
}

func (cw *ChatWindow) loadMessages(chatID uint64, offset int) {
	if cw.loadingMore {
		return
	}

	cw.loadingMore = true

	messages, err := getChatMessages(cw.tokens.AccessToken, chatID, messagesLimit, offset)
	if err != nil {
		cw.showError("Ошибка загрузки сообщений", err)
		cw.loadingMore = false

		return
	}

	messagesMu.Lock()
	if offset == 0 {
		slices.Reverse(messages)
		cw.messages = messages
	}
	messagesMu.Unlock()

	fyne.Do(func() {
		cw.messagesList.Refresh()

		if offset == 0 && len(cw.messages) > 0 {
			cw.messagesList.ScrollToBottom()
		}

		cw.loadingMore = false

		// Если загружено меньше лимита, значит больше нет сообщений
		if len(messages) < messagesLimit {
			cw.hasMore = false
			if cw.loadMoreBtn != nil {
				cw.loadMoreBtn.Hide()
			}
		}
	})
}

func (cw *ChatWindow) sendMessage() {
	text := strings.TrimSpace(cw.messageEntry.Text)
	if text == "" || cw.currentChat == nil {
		return
	}

	localMsg := Message{
		ID:        uint64(time.Now().UnixNano()),
		ChatID:    cw.currentChat.ID,
		Sender:    *currentUser,
		Text:      text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	messagesMu.Lock()

	cw.messages = append(cw.messages, localMsg)
	messagesMu.Unlock()

	cw.messageEntry.SetText("")
	cw.messagesList.Refresh()
	cw.messagesList.ScrollToBottom()

	go func() {
		err := sendWSMessage(cw.currentChat.ID, text)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Не удалось отправить сообщение: %w", err), cw.window)
			})
		}
	}()
}

func (cw *ChatWindow) onNewMessage(msg *Message) {
	if msg.Sender.ID == currentUser.ID {
		return
	}

	// Проверяем, есть ли чат в списке
	chatExists := false

	cw.chatsMu.RLock()

	for _, chat := range cw.chats {
		if chat.ID == msg.ChatID {
			chatExists = true

			break
		}
	}

	cw.chatsMu.RUnlock()

	// Если чата нет в списке - обновляем весь список
	if !chatExists {
		go cw.refreshChats()
	} else {
		// Если чат есть и это НЕ текущий чат - помечаем как непрочитанный
		if cw.currentChat == nil || msg.ChatID != cw.currentChat.ID {
			cw.chatsMu.Lock()
			for i, chat := range cw.chats {
				if chat.ID == msg.ChatID {
					cw.chats[i].IsRead = false

					break
				}
			}
			cw.chatsMu.Unlock()

			fyne.Do(func() {
				cw.chatsList.Refresh()
			})
		}
	}

	// Если сообщение для текущего чата - добавляем его
	if cw.currentChat != nil && msg.ChatID == cw.currentChat.ID {
		messagesMu.Lock()

		cw.messages = append(cw.messages, *msg)
		messagesMu.Unlock()

		fyne.Do(func() {
			cw.messagesList.Refresh()
			cw.messagesList.ScrollToBottom()
		})

		// Помечаем чат как прочитанный
		cw.markChatAsRead(msg.ChatID)
	} else if msg.Sender.ID != currentUser.ID {
		// Показываем уведомление для сообщений из других чатов
		chatTitle := cw.getChatTitle(msg.ChatID)
		showNotification(cw.app, "Новое сообщение", fmt.Sprintf("[%s] %s: %s",
			chatTitle, msg.Sender.Username, msg.Text))
	}
}

func (cw *ChatWindow) getChatTitle(chatID uint64) string {
	cw.chatsMu.RLock()
	defer cw.chatsMu.RUnlock()

	for _, chat := range cw.chats {
		if chat.ID == chatID {
			if chat.Title != "" {
				return chat.Title
			}

			if len(chat.Members) > 0 {
				for _, m := range chat.Members {
					if m.ID != currentUser.ID {
						return m.Username
					}
				}
			}

			return fmt.Sprintf("Чат #%d", chat.ID)
		}
	}

	return fmt.Sprintf("Чат #%d", chatID)
}

func (cw *ChatWindow) logout() {
	go func() {
		// Останавливаем автообновление
		if refreshTicker != nil {
			refreshTicker.Stop()
		}

		if refreshDone != nil {
			close(refreshDone) // Закрываем канал
		}

		// Закрываем WebSocket
		disconnectWebSocket()

		// Выходим из системы на сервере
		_ = logout(cw.tokens.AccessToken)

		// Удаляем файл с токенами
		_ = os.Remove(tokenFile)

		fyne.Do(func() {
			// Закрываем текущее окно чата
			cw.window.Close()

			// Перезапускаем процесс аутентификации
			go authenticateAndRun(cw.app)
		})
	}()
}

func (cw *ChatWindow) showError(title string, err error) {
	fyne.Do(func() {
		dialog.ShowError(fmt.Errorf("%s: %w", title, err), cw.window)
	})
}

// --- Уведомления ---

func showNotification(app fyne.App, title, message string) {
	fyne.Do(func() {
		notifWin := app.NewWindow("")
		notifWin.Resize(fyne.NewSize(300, 100))
		notifWin.SetContent(container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(message),
		))
		notifWin.CenterOnScreen()
		notifWin.Show()

		time.AfterFunc(3*time.Second, func() {
			fyne.Do(func() {
				notifWin.Close()
			})
		})
	})
}

// --- Диалог создания чата ---

func showCreateChatDialog(cw *ChatWindow) {
	win := cw.app.NewWindow("Создать чат")
	win.Resize(fyne.NewSize(450, 400))

	typeEntry := widget.NewSelect([]string{"private", "group"}, nil)
	typeEntry.SetSelected("private")

	membersEntry := widget.NewEntry()
	membersEntry.SetPlaceHolder("ID участников через запятую (например: 1,2,3)")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск пользователей...")

	usersList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", func(bool) {}),
				widget.NewLabel("username"),
				widget.NewLabel("(email)"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)

	selectedUsers := make(map[uint64]bool)

	var users []User

	status := widget.NewLabel("")

	searchBtn := widget.NewButton("Найти", func() {
		query := searchEntry.Text
		if query == "" {
			return
		}

		go func() {
			foundUsers, err := searchUsers(cw.tokens.AccessToken, query, 20, 0)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, win)
				})

				return
			}

			users = foundUsers

			fyne.Do(func() {
				usersList.Length = func() int { return len(users) }
				usersList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
					container_ := item.(*fyne.Container)
					check := container_.Objects[0].(*widget.Check)
					username := container_.Objects[1].(*widget.Label)
					email := container_.Objects[2].(*widget.Label)

					user := users[id]
					username.SetText(user.Username)
					email.SetText("(" + user.Email + ")")

					check.SetChecked(selectedUsers[user.ID])

					check.OnChanged = func(checked bool) {
						if checked {
							selectedUsers[user.ID] = true
						} else {
							delete(selectedUsers, user.ID)
						}
					}
				}
				usersList.Refresh()
			})
		}()
	})

	createBtn := widget.NewButton("Создать", nil)
	createBtn.OnTapped = func() {
		var memberIDs []uint64

		for part := range strings.SplitSeq(membersEntry.Text, ",") {
			if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil {
				memberIDs = append(memberIDs, id)
			}
		}

		for id := range selectedUsers {
			memberIDs = append(memberIDs, id)
		}

		memberIDs = unique(memberIDs)

		if len(memberIDs) == 0 {
			status.SetText("Укажите хотя бы одного участника")

			return
		}

		createBtn.Disable()
		status.SetText("Создание чата...")

		go func() {
			chat, err := createChat(cw.tokens.AccessToken, typeEntry.Selected, memberIDs)
			if err != nil {
				fyne.Do(func() {
					createBtn.Enable()
					status.SetText("Ошибка: " + err.Error())
				})

				return
			}

			fyne.Do(func() {
				status.SetText("Чат создан!")

				// Обновляем список чатов
				cw.refreshChatsAfterNewChat(chat)

				// Открываем созданный чат
				cw.selectChat(chat)

				time.AfterFunc(1*time.Second, func() {
					fyne.Do(func() {
						win.Close()
					})
				})
			})
		}()
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Поиск", container.NewVBox(
			searchEntry,
			searchBtn,
			widget.NewLabel("Выберите пользователей:"),
			container.NewMax(usersList),
		)),
		container.NewTabItem("Ручной ввод", container.NewVBox(
			widget.NewLabel("Введите ID участников:"),
			membersEntry,
		)),
	)

	win.SetContent(container.NewVBox(
		widget.NewLabel("Тип чата:"),
		typeEntry,
		tabs,
		createBtn,
		status,
	))

	win.Show()
}

func showSearchUsersDialog(cw *ChatWindow) {
	win := cw.app.NewWindow("Поиск пользователей")
	win.Resize(fyne.NewSize(400, 500))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Введите имя пользователя...")

	usersList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("username"),
				widget.NewLabel("email"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)

	statusLabel := widget.NewLabel("")

	searchEntry.OnSubmitted = func(s string) {
		if s == "" {
			return
		}

		statusLabel.SetText("Поиск...")

		usersList.Length = func() int { return 0 }
		usersList.Refresh()

		go func() {
			users, err := searchUsers(cw.tokens.AccessToken, s, 50, 0)
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Ошибка: " + err.Error())
				})

				return
			}

			fyne.Do(func() {
				statusLabel.SetText(fmt.Sprintf("Найдено пользователей: %d", len(users)))

				usersList.Length = func() int { return len(users) }
				usersList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
					container_ := item.(*fyne.Container)
					username := container_.Objects[1].(*widget.Label)
					email := container_.Objects[2].(*widget.Label)

					user := users[id]
					username.SetText(user.Username)
					email.SetText(user.Email)
				}
				usersList.Refresh()
			})
		}()
	}

	win.SetContent(container.NewVBox(
		widget.NewLabel("Поиск пользователей"),
		searchEntry,
		statusLabel,
		widget.NewLabel("Результаты:"),
		container.NewMax(usersList),
	))

	win.Show()
}

// --- Окно авторизации ---

func showAuthWindow(myApp fyne.App) {
	authWin := myApp.NewWindow("Вход / Регистрация")
	authWin.Resize(fyne.NewSize(400, 450))

	// Устанавливаем обработчик закрытия окна авторизации
	authWin.SetCloseIntercept(func() {
		shutdown(myApp)
	})

	// Создаем вкладки
	tabs := container.NewAppTabs(
		container.NewTabItem("Вход", container.NewVBox()),
		container.NewTabItem("Регистрация", container.NewVBox()),
	)

	// Вход
	loginEmail := widget.NewEntry()
	loginEmail.SetPlaceHolder("Email")

	loginPassword := widget.NewPasswordEntry()
	loginPassword.SetPlaceHolder("Пароль")

	// Регистрация
	registerEmail := widget.NewEntry()
	registerEmail.SetPlaceHolder("Email")

	registerUsername := widget.NewEntry()
	registerUsername.SetPlaceHolder("Имя пользователя (от 5 символов)")

	registerPassword := widget.NewPasswordEntry()
	registerPassword.SetPlaceHolder("Пароль")

	registerConfirm := widget.NewPasswordEntry()
	registerConfirm.SetPlaceHolder("Подтверждение пароля")

	statusLabel := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hidden = true

	// Логин
	loginBtn := widget.NewButton("Войти", func() {
		if loginEmail.Text == "" || loginPassword.Text == "" {
			statusLabel.SetText("Заполните все поля")

			return
		}

		progress.Hidden = false

		statusLabel.SetText("")

		go func() {
			tokens, err := login(loginEmail.Text, loginPassword.Text)
			if err != nil {
				fyne.Do(func() {
					progress.Hidden = true

					statusLabel.SetText("Ошибка: " + err.Error())
				})

				return
			}

			user, err := getCurrentUser(tokens.AccessToken)
			if err != nil {
				fyne.Do(func() {
					progress.Hidden = true

					statusLabel.SetText("Ошибка получения пользователя: " + err.Error())
				})

				return
			}

			currentUser = user
			currentTokens = tokens
			saveTokensToFile(tokens)
			startTokenRefreshRoutine()

			fyne.Do(func() {
				authWin.Close()
				showMainChatWindow(myApp, tokens)
			})
		}()
	})

	loginTab := container.NewVBox(
		widget.NewLabelWithStyle("Вход", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		loginEmail,
		loginPassword,
		loginBtn,
	)

	// Регистрация
	registerBtn := widget.NewButton("Зарегистрироваться", func() {
		// Валидация
		if registerEmail.Text == "" || registerUsername.Text == "" ||
			registerPassword.Text == "" || registerConfirm.Text == "" {
			statusLabel.SetText("Заполните все поля")

			return
		}

		if registerPassword.Text != registerConfirm.Text {
			statusLabel.SetText("Пароли не совпадают")

			return
		}

		if len(registerUsername.Text) < 5 {
			statusLabel.SetText("Имя пользователя должно быть не менее 5 символов")

			return
		}

		progress.Hidden = false

		statusLabel.SetText("Регистрация...")

		go func() {
			// Вызываем только регистрацию, без автоматического логина
			err := registerUser(registerEmail.Text, registerUsername.Text, registerPassword.Text)
			if err != nil {
				fyne.Do(func() {
					progress.Hidden = true

					statusLabel.SetText("Ошибка регистрации: " + err.Error())
				})

				return
			}

			fyne.Do(func() {
				progress.Hidden = true

				statusLabel.SetText("Регистрация успешна! Теперь войдите.")

				// Очищаем поля регистрации
				registerEmail.SetText("")
				registerUsername.SetText("")
				registerPassword.SetText("")
				registerConfirm.SetText("")

				// Заполняем поля входа тем же email
				loginEmail.SetText(registerEmail.Text)
				loginPassword.SetText("")

				// Переключаемся на вкладку входа
				tabs.SelectIndex(0)

				// Убираем сообщение об успехе через 3 секунды
				time.AfterFunc(3*time.Second, func() {
					fyne.Do(func() {
						statusLabel.SetText("")
					})
				})
			})
		}()
	})

	registerTab := container.NewVBox(
		widget.NewLabelWithStyle("Регистрация", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		registerEmail,
		registerUsername,
		registerPassword,
		registerConfirm,
		registerBtn,
	)

	// Обновляем содержимое вкладок
	tabs.Items[0].Content = loginTab
	tabs.Items[1].Content = registerTab

	content := container.NewVBox(
		tabs,
		widget.NewSeparator(),
		progress,
		statusLabel,
	)

	authWin.SetContent(content)
	authWin.Show()
}

// Новая функция только для регистрации (без автоматического логина).
func registerUser(email, username, password string) error {
	input := RegisterInput{
		Email:    email,
		Username: username,
		Password: password,
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("registration failed (%s): %s", resp.Status, string(errorBody))
	}

	return nil
}

// --- Вспомогательные функции ---

func unique(ids []uint64) []uint64 {
	seen := make(map[uint64]bool)
	result := []uint64{}

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true

			result = append(result, id)
		}
	}

	return result
}

// --- Функция shutdown (добавьте после main) ---.
func shutdown(app fyne.App) {
	fmt.Println("Завершение приложения...")

	// Используем Once, чтобы закрыть канал только один раз
	shutdownOnce.Do(func() {
		// Сигнализируем всем горутинам о завершении
		close(appShutdown)

		// Останавливаем автообновление токенов
		if refreshTicker != nil {
			refreshTicker.Stop()
		}

		// Безопасно закрываем refreshDone
		refreshMu.Lock()
		if refreshDone != nil {
			// Проверяем, не закрыт ли уже канал
			select {
			case <-refreshDone:
				// Канал уже закрыт, ничего не делаем
			default:
				close(refreshDone)
			}
		}
		refreshMu.Unlock()

		// Закрываем WebSocket соединение
		disconnectWebSocket()

		// Ждем завершения всех горутин (с таймаутом)
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("Все горутины завершены")
		case <-time.After(3 * time.Second):
			fmt.Println("Таймаут ожидания горутин (принудительное завершение)")
		}
	})

	// Завершаем приложение
	app.Quit()
}

// --- Функция isShuttingDown (добавьте после shutdown) ---.
func isShuttingDown() bool {
	select {
	case <-appShutdown:
		return true
	default:
		return false
	}
}
