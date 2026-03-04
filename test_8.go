package main

//
//import (
//	"bytes"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"io"
//	"kfcGUI/internal/common"
//	"net/http"
//	"net/url"
//	"os"
//	"slices"
//	"strconv"
//	"strings"
//	"sync"
//	"time"
//
//	"fyne.io/fyne/v2"
//	"fyne.io/fyne/v2/app"
//	"fyne.io/fyne/v2/container"
//	"fyne.io/fyne/v2/dialog"
//	"fyne.io/fyne/v2/theme"
//	"fyne.io/fyne/v2/widget"
//	"github.com/gorilla/websocket"
//)
//
//const (
//	tokenFile       = "tokens.json" // Теперь храним оба токена в JSON
//	apiBase         = "http://localhost:8080"
//	wsBase          = "ws://localhost:8080"
//	messagesLimit   = 10
//	chatsLimit      = 20
//	refreshInterval = 5 * time.Minute // Обновление каждые 5 минут
//)
//
//// --- Структуры для токенов ---
//
//type TokenPair struct {
//	AccessToken  string    `json:"accessToken"`
//	RefreshToken string    `json:"refreshToken"`
//	ExpiresAt    time.Time `json:"expiresAt"` // Когда нужно обновить
//}
//
//// --- Структуры данных из Swagger ---
//
//type User struct {
//	ID             uint64    `json:"id"`
//	Username       string    `json:"username"`
//	Email          string    `json:"email"`
//	EmailConfirmed bool      `json:"emailConfirmed"`
//	CreatedAt      time.Time `json:"createdAt"`
//	UpdatedAt      time.Time `json:"updatedAt"`
//}
//
//type Sender struct {
//	ID       uint64 `json:"id"`
//	Username string `json:"username"`
//}
//
//type Message struct {
//	ID        uint64    `json:"id"`
//	ChatID    uint64    `json:"chatId"`
//	Sender    User      `json:"sender"`
//	Text      string    `json:"text"`
//	CreatedAt time.Time `json:"createdAt"`
//	UpdatedAt time.Time `json:"updatedAt"`
//}
//
//type OutgoingMessage struct {
//	ChatID uint64 `json:"chatId"`
//	Sender User   `json:"sender"`
//	Text   string `json:"text"`
//}
//
//type Chat struct {
//	ID          uint64    `json:"id"`
//	Title       string    `json:"title,omitempty"`
//	Description string    `json:"description,omitempty"`
//	Type        string    `json:"type"`
//	Members     []User    `json:"members,omitempty"`
//	Messages    []Message `json:"messages,omitempty"`
//	IsRead      bool      `json:"isRead"`
//	CreatedAt   time.Time `json:"createdAt"`
//	UpdatedAt   time.Time `json:"updatedAt"`
//}
//
//type MemberInput struct {
//	ID uint64 `json:"id"`
//}
//
//type LoginInput struct {
//	Email    string `json:"email"`
//	Password string `json:"password"`
//}
//
//type RegisterInput struct {
//	Email    string `json:"email"`
//	Username string `json:"username"`
//	Password string `json:"password"`
//}
//
//// --- Глобальные переменные ---
//var (
//	httpClient    = &http.Client{Timeout: 10 * time.Second}
//	currentUser   *User
//	currentTokens *TokenPair
//	wsConn        *websocket.Conn
//	wsMutex       sync.Mutex
//	chatMessages  = make(map[uint64][]Message)
//	messagesMu    sync.RWMutex
//	refreshTicker *time.Ticker
//	refreshDone   chan bool
//)
//
//// --- Основная функция ---
//
//func main() {
//	myApp := app.New()
//	myApp.Settings().SetTheme(theme.LightTheme())
//
//	mainWindow := myApp.NewWindow("KFC Chat")
//	mainWindow.Resize(fyne.NewSize(400, 300))
//	mainWindow.SetContent(container.NewCenter(
//		container.NewVBox(
//			widget.NewProgressBarInfinite(),
//			widget.NewLabel("Проверка авторизации..."),
//		),
//	))
//
//	// Запускаем асинхронную проверку
//	go authenticateAndRun(myApp, mainWindow)
//
//	mainWindow.Show()
//	myApp.Run()
//
//	// Останавливаем автообновление при выходе
//	if refreshTicker != nil {
//		refreshTicker.Stop()
//		refreshDone <- true
//	}
//}
//
//// --- Аутентификация ---
//
//func authenticateAndRun(myApp fyne.App, mainWindow fyne.Window) {
//	tokens, err := readTokensFromFile()
//
//	if err == nil && tokens != nil && tokens.AccessToken != "" && tokens.RefreshToken != "" {
//		currentTokens = tokens
//
//		// Пробуем получить пользователя с текущим accessToken
//		user, err := getCurrentUser(tokens.AccessToken)
//		if err == nil && user != nil {
//			// Успешная авторизация
//			currentUser = user
//			startTokenRefreshRoutine(myApp)
//
//			fyne.Do(func() {
//				mainWindow.Close()
//				showMainChatWindow(myApp, tokens)
//			})
//			return
//		}
//
//		// Если accessToken не работает, пробуем обновить токены
//		newTokens, refreshErr := refreshTokens(tokens.RefreshToken)
//		if refreshErr == nil && newTokens != nil {
//			// Сохраняем новые токены
//			saveTokensToFile(newTokens)
//			currentTokens = newTokens
//
//			// Получаем пользователя с новым токеном
//			user, err := getCurrentUser(newTokens.AccessToken)
//			if err == nil && user != nil {
//				currentUser = user
//				startTokenRefreshRoutine(myApp)
//
//				fyne.Do(func() {
//					mainWindow.Close()
//					showMainChatWindow(myApp, newTokens)
//				})
//				return
//			}
//		}
//	}
//
//	// Если дошли сюда - нужна авторизация
//	fyne.Do(func() {
//		showAuthWindow(myApp, mainWindow)
//	})
//}
//
//// --- Работа с токенами ---
//
//func readTokensFromFile() (*TokenPair, error) {
//	data, err := os.ReadFile(tokenFile)
//	if err != nil {
//		return nil, err
//	}
//
//	var tokens TokenPair
//	if err := json.Unmarshal(data, &tokens); err != nil {
//		return nil, err
//	}
//
//	return &tokens, nil
//}
//
//func saveTokensToFile(tokens *TokenPair) error {
//	// Устанавливаем время истечения (например, через 4 минуты 30 секунд для надежности)
//	tokens.ExpiresAt = time.Now().Add(4*time.Minute + 30*time.Second)
//
//	data, err := json.MarshalIndent(tokens, "", "  ")
//	if err != nil {
//		return err
//	}
//
//	return os.WriteFile(tokenFile, data, 0600)
//}
//
//func startTokenRefreshRoutine(myApp fyne.App) {
//	refreshTicker = time.NewTicker(refreshInterval)
//	refreshDone = make(chan bool)
//
//	go func() {
//		for {
//			select {
//			case <-refreshDone:
//				return
//			case <-refreshTicker.C:
//				refreshTokenRoutine(myApp)
//			}
//		}
//	}()
//}
//
//func refreshTokenRoutine(myApp fyne.App) {
//	if currentTokens == nil || currentTokens.RefreshToken == "" {
//		return
//	}
//
//	newTokens, err := refreshTokens(currentTokens.RefreshToken)
//	if err != nil {
//		fmt.Printf("Ошибка обновления токенов: %v\n", err)
//		return
//	}
//
//	// Обновляем текущие токены
//	currentTokens = newTokens
//	saveTokensToFile(newTokens)
//
//	fmt.Println("Токены успешно обновлены")
//}
//
//// --- HTTP запросы к API ---
//
//func getCurrentUser(accessToken string) (*User, error) {
//	req, _ := http.NewRequest("GET", apiBase+"/users/me", nil)
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("API error: %s", resp.Status)
//	}
//
//	var user User
//	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
//		return nil, err
//	}
//	return &user, nil
//}
//
//func refreshTokens(refreshToken string) (*TokenPair, error) {
//	req, _ := http.NewRequest("PUT", apiBase+"/sessions", nil)
//	req.Header.Set("Cookie", "refreshToken="+refreshToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusNoContent {
//		return nil, fmt.Errorf("refresh failed: %s", resp.Status)
//	}
//
//	tokens := &TokenPair{}
//
//	// Извлекаем токены из cookie
//	for _, cookie := range resp.Cookies() {
//		if cookie.Name == "accessToken" {
//			tokens.AccessToken = cookie.Value
//		}
//		if cookie.Name == "refreshToken" {
//			tokens.RefreshToken = cookie.Value
//		}
//	}
//
//	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
//		return nil, errors.New("не удалось получить токены из cookie")
//	}
//
//	return tokens, nil
//}
//
//func login(email, password string) (*TokenPair, error) {
//	input := LoginInput{
//		Email:    email,
//		Password: password,
//	}
//	body, _ := json.Marshal(input)
//
//	req, _ := http.NewRequest("POST", apiBase+"/sessions", bytes.NewReader(body))
//	req.Header.Set("Content-Type", "application/json")
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusNoContent {
//		errorBody, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("login failed (%s): %s", resp.Status, string(errorBody))
//	}
//
//	tokens := &TokenPair{}
//
//	// Извлекаем токены из cookie
//	for _, cookie := range resp.Cookies() {
//		if cookie.Name == "accessToken" {
//			tokens.AccessToken = cookie.Value
//		}
//		if cookie.Name == "refreshToken" {
//			tokens.RefreshToken = cookie.Value
//		}
//	}
//
//	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
//		return nil, errors.New("не удалось получить токены из cookie")
//	}
//
//	return tokens, nil
//}
//
//func register(email, username, password string) (*TokenPair, error) {
//	input := RegisterInput{
//		Email:    email,
//		Username: username,
//		Password: password,
//	}
//	body, _ := json.Marshal(input)
//
//	req, _ := http.NewRequest("POST", apiBase+"/users", bytes.NewReader(body))
//	req.Header.Set("Content-Type", "application/json")
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusCreated {
//		errorBody, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("registration failed (%s): %s", resp.Status, string(errorBody))
//	}
//
//	// После регистрации автоматически логинимся
//	return login(email, password)
//}
//
//func logout(accessToken string) error {
//	req, _ := http.NewRequest("DELETE", apiBase+"/sessions", nil)
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusNoContent {
//		return fmt.Errorf("logout failed: %s", resp.Status)
//	}
//	return nil
//}
//
//func getUserChats(accessToken string, limit, offset int) ([]Chat, error) {
//	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/chats?limit=%d&offset=%d", apiBase, limit, offset), nil)
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("API error: %s", resp.Status)
//	}
//
//	var chats []Chat
//	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
//		return nil, err
//	}
//	return chats, nil
//}
//
//func getChatMessages(accessToken string, chatID uint64, limit, offset int) ([]Message, error) {
//	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/chats/%d/messages?limit=%d&offset=%d", apiBase, chatID, limit, offset), nil)
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("API error: %s", resp.Status)
//	}
//
//	var messages []Message
//	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
//		return nil, err
//	}
//	return messages, nil
//}
//
//func createChat(accessToken string, chatType string, memberIDs []uint64) (*Chat, error) {
//	members := make([]MemberInput, len(memberIDs))
//	for i, id := range memberIDs {
//		members[i] = MemberInput{ID: id}
//	}
//
//	bodyMap := map[string]interface{}{
//		"type":    chatType,
//		"members": members,
//	}
//	body, _ := json.Marshal(bodyMap)
//
//	req, _ := http.NewRequest("POST", apiBase+"/chats", bytes.NewReader(body))
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//	req.Header.Set("Content-Type", "application/json")
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusCreated {
//		errorBody, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(errorBody))
//	}
//
//	var chat Chat
//	if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
//		return nil, err
//	}
//	return &chat, nil
//}
//
//func searchUsers(accessToken, query string, limit, offset int) ([]User, error) {
//	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/users?username=%s&limit=%d&offset=%d",
//		apiBase, url.QueryEscape(query), limit, offset), nil)
//	req.Header.Set("Cookie", "accessToken="+accessToken)
//
//	resp, err := httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("API error: %s", resp.Status)
//	}
//
//	var users []User
//	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
//		return nil, err
//	}
//	return users, nil
//}
//
//// --- WebSocket подключение ---
//
//func connectWebSocket(accessToken string, chatID uint64, onMessage func(*Message)) error {
//	wsMutex.Lock()
//	defer wsMutex.Unlock()
//
//	if wsConn != nil {
//		return nil
//	}
//
//	header := http.Header{}
//	header.Add("Cookie", "accessToken="+accessToken)
//
//	conn, _, err := websocket.DefaultDialer.Dial(wsBase+"/ws", header)
//	if err != nil {
//		return err
//	}
//
//	wsConn = conn
//
//	// Запускаем чтение сообщений
//	go func() {
//		for {
//			var msg Message
//			err := wsConn.ReadJSON(&msg)
//			if err != nil {
//				wsMutex.Lock()
//				wsConn.Close()
//				wsConn = nil
//				wsMutex.Unlock()
//
//				// Пытаемся переподключиться с новым токеном
//				if currentTokens != nil {
//					time.Sleep(3 * time.Second)
//					connectWebSocket(currentTokens.AccessToken, chatID, onMessage)
//				}
//				return
//			}
//
//			// Вызываем колбэк с полученным сообщением
//			onMessage(&msg)
//		}
//	}()
//
//	return nil
//}
//
//func sendWSMessage(chatID uint64, text string) error {
//	wsMutex.Lock()
//	defer wsMutex.Unlock()
//
//	if wsConn == nil {
//		return errors.New("websocket not connected")
//	}
//
//	outgoingMsg := OutgoingMessage{
//		ChatID: chatID,
//		Sender: *currentUser,
//		Text:   text,
//	}
//
//	return wsConn.WriteJSON(outgoingMsg)
//}
//
//func disconnectWebSocket() {
//	wsMutex.Lock()
//	defer wsMutex.Unlock()
//
//	if wsConn != nil {
//		wsConn.Close()
//		wsConn = nil
//	}
//}
//
//// --- Основное окно чата ---
//
//type ChatWindow struct {
//	app          fyne.App
//	tokens       *TokenPair
//	window       fyne.Window
//	chatsList    *widget.List
//	messagesList *widget.List
//	messageEntry *widget.Entry
//	currentChat  *Chat
//	chats        []Chat
//	messages     []Message
//	statusLabel  *widget.Label
//	loadingMore  bool
//}
//
//func showMainChatWindow(myApp fyne.App, tokens *TokenPair) {
//	cw := &ChatWindow{
//		app:    myApp,
//		tokens: tokens,
//		window: myApp.NewWindow("KFC Chat - " + currentUser.Username),
//	}
//
//	cw.window.Resize(fyne.NewSize(900, 700))
//	cw.window.SetCloseIntercept(func() {
//		// Останавливаем автообновление при закрытии окна
//		if refreshTicker != nil {
//			refreshTicker.Stop()
//			refreshDone <- true
//		}
//		cw.window.Close()
//	})
//
//	// Статусная строка
//	cw.statusLabel = widget.NewLabel("")
//
//	// Список чатов
//	cw.chatsList = widget.NewList(
//		func() int { return len(cw.chats) },
//		func() fyne.CanvasObject {
//			return container.NewHBox(
//				widget.NewIcon(theme.MailComposeIcon()),
//				widget.NewLabel("Chat"),
//			)
//		},
//		func(id widget.ListItemID, item fyne.CanvasObject) {
//			if id >= len(cw.chats) {
//				return
//			}
//			chat := cw.chats[id]
//			container := item.(*fyne.Container)
//			label := container.Objects[1].(*widget.Label)
//
//			title := chat.Title
//			if title == "" && len(chat.Members) > 0 {
//				for _, m := range chat.Members {
//					if m.ID != currentUser.ID {
//						title = m.Username
//						break
//					}
//				}
//			}
//			if title == "" {
//				title = fmt.Sprintf("Chat #%d", chat.ID)
//			}
//
//			label.SetText(title)
//			if !chat.IsRead {
//				label.TextStyle = fyne.TextStyle{Bold: true}
//			} else {
//				label.TextStyle = fyne.TextStyle{}
//			}
//		},
//	)
//
//	cw.chatsList.OnSelected = func(id widget.ListItemID) {
//		if id < len(cw.chats) {
//			cw.selectChat(&cw.chats[id])
//		}
//	}
//
//	// Список сообщений
//	cw.messagesList = widget.NewList(
//		func() int { return len(cw.messages) },
//		func() fyne.CanvasObject {
//			// Контейнер для сообщения
//			sender := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
//			message := widget.NewLabel("")
//			message.Wrapping = fyne.TextWrapWord
//			time_ := widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})
//
//			return container.NewVBox(
//				container.NewBorder(nil, nil, sender, time_),
//				message,
//				widget.NewSeparator(),
//			)
//		},
//		func(id widget.ListItemID, item fyne.CanvasObject) {
//			if id >= len(cw.messages) {
//				return
//			}
//
//			msg := cw.messages[id]
//
//			container_ := item.(*fyne.Container)
//			border := container_.Objects[0].(*fyne.Container)
//			sender := border.Objects[0].(*widget.Label)
//			time_ := border.Objects[1].(*widget.Label)
//			message := container_.Objects[1].(*widget.Label)
//
//			if msg.Sender.ID == currentUser.ID {
//				sender.SetText("Вы")
//			} else {
//				sender.SetText(msg.Sender.Username)
//			}
//			message.SetText(msg.Text)
//			time_.SetText(msg.CreatedAt.Format(common.DateTimeFormat))
//		},
//	)
//
//	// Поле ввода
//	cw.messageEntry = widget.NewMultiLineEntry()
//	cw.messageEntry.SetPlaceHolder("Введите сообщение...")
//	cw.messageEntry.OnSubmitted = func(s string) {
//		cw.sendMessage()
//	}
//
//	sendBtn := widget.NewButtonWithIcon("Отправить", theme.MailSendIcon(), func() {
//		cw.sendMessage()
//	})
//
//	// Кнопки управления
//	newChatBtn := widget.NewButtonWithIcon("Новый чат", theme.ContentAddIcon(), func() {
//		showCreateChatDialog(cw)
//	})
//
//	searchBtn := widget.NewButtonWithIcon("Поиск", theme.SearchIcon(), func() {
//		showSearchUsersDialog(cw)
//	})
//
//	logoutBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
//		cw.logout()
//	})
//
//	// Сборка интерфейса
//	toolbar := container.NewHBox(newChatBtn, searchBtn, logoutBtn)
//
//	leftPanel := container.NewBorder(
//		container.NewVBox(
//			widget.NewLabelWithStyle("Чаты", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//			widget.NewSeparator(),
//		),
//		cw.statusLabel,
//		nil, nil,
//		container.NewBorder(
//			toolbar,
//			nil, nil, nil,
//			cw.chatsList,
//		),
//	)
//
//	inputArea := container.NewBorder(nil, nil, nil, sendBtn, cw.messageEntry)
//
//	rightPanel := container.NewBorder(
//		widget.NewLabelWithStyle("Сообщения", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//		inputArea,
//		nil, nil,
//		container.NewBorder(
//			nil, nil, nil, nil,
//			cw.messagesList,
//		),
//	)
//
//	split := container.NewHSplit(leftPanel, rightPanel)
//	split.Offset = 0.25
//
//	cw.window.SetContent(split)
//
//	// Загружаем чаты
//	go cw.loadChats()
//
//	cw.window.Show()
//}
//
//func (cw *ChatWindow) loadChats() {
//	chats, err := getUserChats(cw.tokens.AccessToken, chatsLimit, 0)
//	if err != nil {
//		cw.showError("Ошибка загрузки чатов", err)
//		return
//	}
//
//	cw.chats = chats
//	fyne.Do(func() {
//		cw.chatsList.Refresh()
//	})
//}
//
//func (cw *ChatWindow) selectChat(chat *Chat) {
//	cw.currentChat = chat
//	cw.messages = nil
//	cw.messagesList.Refresh()
//
//	// Загружаем сообщения
//	go cw.loadMessages(chat.ID, 0)
//
//	// Подключаем WebSocket
//	go func() {
//		err := connectWebSocket(cw.tokens.AccessToken, chat.ID, func(msg *Message) {
//			cw.onNewMessage(msg)
//		})
//		if err != nil {
//			cw.showError("Ошибка WebSocket", err)
//		}
//	}()
//}
//
//func (cw *ChatWindow) loadMessages(chatID uint64, offset int) {
//	if cw.loadingMore {
//		return
//	}
//	cw.loadingMore = true
//
//	messages, err := getChatMessages(cw.tokens.AccessToken, chatID, messagesLimit, offset)
//	if err != nil {
//		cw.showError("Ошибка загрузки сообщений", err)
//		cw.loadingMore = false
//		return
//	}
//
//	messagesMu.Lock()
//	if offset == 0 {
//		slices.Reverse(messages)
//		cw.messages = messages
//	} else {
//		cw.messages = append(messages, cw.messages...)
//	}
//	messagesMu.Unlock()
//
//	fyne.Do(func() {
//		cw.messagesList.Refresh()
//		if offset == 0 && len(cw.messages) > 0 {
//			cw.messagesList.ScrollToBottom()
//		}
//	})
//
//	cw.loadingMore = false
//}
//
//func (cw *ChatWindow) sendMessage() {
//	text := strings.TrimSpace(cw.messageEntry.Text)
//	if text == "" || cw.currentChat == nil {
//		return
//	}
//
//	localMsg := Message{
//		ID:        uint64(time.Now().UnixNano()),
//		ChatID:    cw.currentChat.ID,
//		Sender:    *currentUser,
//		Text:      text,
//		CreatedAt: time.Now().In(common.Timezone),
//		UpdatedAt: time.Now().In(common.Timezone),
//	}
//
//	messagesMu.Lock()
//	cw.messages = append(cw.messages, localMsg)
//	messagesMu.Unlock()
//
//	cw.messageEntry.SetText("")
//	cw.messagesList.Refresh()
//	cw.messagesList.ScrollToBottom()
//
//	go func() {
//		err := sendWSMessage(cw.currentChat.ID, text)
//		if err != nil {
//			fyne.Do(func() {
//				dialog.ShowError(fmt.Errorf("Не удалось отправить сообщение: %v", err), cw.window)
//			})
//		}
//	}()
//}
//
//func (cw *ChatWindow) onNewMessage(msg *Message) {
//	if msg.Sender.ID == currentUser.ID {
//		return
//	}
//
//	if cw.currentChat != nil && msg.ChatID == cw.currentChat.ID {
//		messagesMu.Lock()
//		cw.messages = append(cw.messages, *msg)
//		messagesMu.Unlock()
//
//		fyne.Do(func() {
//			cw.messagesList.Refresh()
//			cw.messagesList.ScrollToBottom()
//		})
//	} else {
//		for i, chat := range cw.chats {
//			if chat.ID == msg.ChatID {
//				cw.chats[i].IsRead = false
//				fyne.Do(func() {
//					cw.chatsList.Refresh()
//				})
//				break
//			}
//		}
//	}
//
//	if (cw.currentChat == nil || msg.ChatID != cw.currentChat.ID) && msg.Sender.ID != currentUser.ID {
//		chatTitle := fmt.Sprintf("Чат %d", msg.ChatID)
//		for _, chat := range cw.chats {
//			if chat.ID == msg.ChatID {
//				if chat.Title != "" {
//					chatTitle = chat.Title
//				} else if len(chat.Members) > 0 {
//					for _, m := range chat.Members {
//						if m.ID != currentUser.ID {
//							chatTitle = m.Username
//							break
//						}
//					}
//				}
//				break
//			}
//		}
//		showNotification(cw.app, "Новое сообщение", fmt.Sprintf("[%s] %s: %s",
//			chatTitle, msg.Sender.Username, msg.Text))
//	}
//}
//
//func (cw *ChatWindow) logout() {
//	go func() {
//		// Останавливаем автообновление
//		if refreshTicker != nil {
//			refreshTicker.Stop()
//			refreshDone <- true
//		}
//
//		_ = logout(cw.tokens.AccessToken)
//		_ = os.Remove(tokenFile)
//		disconnectWebSocket()
//
//		fyne.Do(func() {
//			cw.window.Close()
//
//			mainWindow := cw.app.NewWindow("KFC Chat")
//			mainWindow.Resize(fyne.NewSize(400, 300))
//			showAuthWindow(cw.app, mainWindow)
//			mainWindow.Show()
//		})
//	}()
//}
//
//func (cw *ChatWindow) showError(title string, err error) {
//	fyne.Do(func() {
//		dialog.ShowError(fmt.Errorf("%s: %w", title, err), cw.window)
//	})
//}
//
//// --- Уведомления ---
//
//func showNotification(app fyne.App, title, message string) {
//	fyne.Do(func() {
//		notifWin := app.NewWindow("")
//		notifWin.Resize(fyne.NewSize(300, 100))
//		notifWin.SetContent(container.NewVBox(
//			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//			widget.NewLabel(message),
//		))
//		notifWin.CenterOnScreen()
//		notifWin.Show()
//
//		time.AfterFunc(3*time.Second, func() {
//			fyne.Do(func() {
//				notifWin.Close()
//			})
//		})
//	})
//}
//
//// --- Диалог создания чата ---
//
//func showCreateChatDialog(cw *ChatWindow) {
//	win := cw.app.NewWindow("Создать чат")
//	win.Resize(fyne.NewSize(450, 400))
//
//	typeEntry := widget.NewSelect([]string{"private", "group"}, nil)
//	typeEntry.SetSelected("private")
//
//	membersEntry := widget.NewEntry()
//	membersEntry.SetPlaceHolder("ID участников через запятую (например: 1,2,3)")
//
//	searchEntry := widget.NewEntry()
//	searchEntry.SetPlaceHolder("Поиск пользователей...")
//
//	usersList := widget.NewList(
//		func() int { return 0 },
//		func() fyne.CanvasObject {
//			return container.NewHBox(
//				widget.NewCheck("", func(bool) {}),
//				widget.NewLabel("username"),
//				widget.NewLabel("(email)"),
//			)
//		},
//		func(id widget.ListItemID, item fyne.CanvasObject) {},
//	)
//
//	selectedUsers := make(map[uint64]bool)
//	var users []User
//	status := widget.NewLabel("")
//
//	searchBtn := widget.NewButton("Найти", func() {
//		query := searchEntry.Text
//		if query == "" {
//			return
//		}
//
//		go func() {
//			foundUsers, err := searchUsers(cw.tokens.AccessToken, query, 20, 0)
//			if err != nil {
//				fyne.Do(func() {
//					dialog.ShowError(err, win)
//				})
//				return
//			}
//
//			users = foundUsers
//
//			fyne.Do(func() {
//				usersList.Length = func() int { return len(users) }
//				usersList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
//					container_ := item.(*fyne.Container)
//					check := container_.Objects[0].(*widget.Check)
//					username := container_.Objects[1].(*widget.Label)
//					email := container_.Objects[2].(*widget.Label)
//
//					user := users[id]
//					username.SetText(user.Username)
//					email.SetText("(" + user.Email + ")")
//
//					check.SetChecked(selectedUsers[user.ID])
//
//					check.OnChanged = func(checked bool) {
//						if checked {
//							selectedUsers[user.ID] = true
//						} else {
//							delete(selectedUsers, user.ID)
//						}
//					}
//				}
//				usersList.Refresh()
//			})
//		}()
//	})
//
//	createBtn := widget.NewButton("Создать", nil)
//	createBtn.OnTapped = func() {
//		var memberIDs []uint64
//
//		for _, part := range strings.Split(membersEntry.Text, ",") {
//			if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil {
//				memberIDs = append(memberIDs, id)
//			}
//		}
//
//		for id := range selectedUsers {
//			memberIDs = append(memberIDs, id)
//		}
//
//		memberIDs = unique(memberIDs)
//
//		if len(memberIDs) == 0 {
//			status.SetText("Укажите хотя бы одного участника")
//			return
//		}
//
//		createBtn.Disable()
//		status.SetText("Создание чата...")
//
//		go func() {
//			chat, err := createChat(cw.tokens.AccessToken, typeEntry.Selected, memberIDs)
//			if err != nil {
//				fyne.Do(func() {
//					createBtn.Enable()
//					status.SetText("Ошибка: " + err.Error())
//				})
//				return
//			}
//
//			fyne.Do(func() {
//				status.SetText("Чат создан!")
//				go cw.loadChats()
//
//				if chat != nil {
//					cw.selectChat(chat)
//				}
//
//				time.AfterFunc(1*time.Second, func() {
//					fyne.Do(func() {
//						win.Close()
//					})
//				})
//			})
//		}()
//	}
//
//	tabs := container.NewAppTabs(
//		container.NewTabItem("Поиск", container.NewVBox(
//			searchEntry,
//			searchBtn,
//			widget.NewLabel("Выберите пользователей:"),
//			container.NewMax(usersList),
//		)),
//		container.NewTabItem("Ручной ввод", container.NewVBox(
//			widget.NewLabel("Введите ID участников:"),
//			membersEntry,
//		)),
//	)
//
//	win.SetContent(container.NewVBox(
//		widget.NewLabel("Тип чата:"),
//		typeEntry,
//		tabs,
//		createBtn,
//		status,
//	))
//
//	win.Show()
//}
//
//func showSearchUsersDialog(cw *ChatWindow) {
//	win := cw.app.NewWindow("Поиск пользователей")
//	win.Resize(fyne.NewSize(400, 500))
//
//	searchEntry := widget.NewEntry()
//	searchEntry.SetPlaceHolder("Введите имя пользователя...")
//
//	usersList := widget.NewList(
//		func() int { return 0 },
//		func() fyne.CanvasObject {
//			return container.NewHBox(
//				widget.NewIcon(theme.AccountIcon()),
//				widget.NewLabel("username"),
//				widget.NewLabel("email"),
//			)
//		},
//		func(id widget.ListItemID, item fyne.CanvasObject) {},
//	)
//
//	statusLabel := widget.NewLabel("")
//
//	searchEntry.OnSubmitted = func(s string) {
//		if s == "" {
//			return
//		}
//
//		statusLabel.SetText("Поиск...")
//		usersList.Length = func() int { return 0 }
//		usersList.Refresh()
//
//		go func() {
//			users, err := searchUsers(cw.tokens.AccessToken, s, 50, 0)
//			if err != nil {
//				fyne.Do(func() {
//					statusLabel.SetText("Ошибка: " + err.Error())
//				})
//				return
//			}
//
//			fyne.Do(func() {
//				statusLabel.SetText(fmt.Sprintf("Найдено пользователей: %d", len(users)))
//
//				usersList.Length = func() int { return len(users) }
//				usersList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
//					container_ := item.(*fyne.Container)
//					username := container_.Objects[1].(*widget.Label)
//					email := container_.Objects[2].(*widget.Label)
//
//					user := users[id]
//					username.SetText(user.Username)
//					email.SetText(user.Email)
//				}
//				usersList.Refresh()
//			})
//		}()
//	}
//
//	win.SetContent(container.NewVBox(
//		widget.NewLabel("Поиск пользователей"),
//		searchEntry,
//		statusLabel,
//		widget.NewLabel("Результаты:"),
//		container.NewMax(usersList),
//	))
//
//	win.Show()
//}
//
//// --- Окно авторизации ---
//
//func showAuthWindow(myApp fyne.App, parentWindow fyne.Window) {
//	authWin := myApp.NewWindow("Вход / Регистрация")
//	authWin.Resize(fyne.NewSize(400, 400))
//
//	// Вход
//	loginEmail := widget.NewEntry()
//	loginEmail.SetPlaceHolder("Email")
//	loginPassword := widget.NewPasswordEntry()
//	loginPassword.SetPlaceHolder("Пароль")
//
//	// Регистрация
//	registerEmail := widget.NewEntry()
//	registerEmail.SetPlaceHolder("Email")
//	registerUsername := widget.NewEntry()
//	registerUsername.SetPlaceHolder("Имя пользователя (от 5 символов)")
//	registerPassword := widget.NewPasswordEntry()
//	registerPassword.SetPlaceHolder("Пароль")
//	registerConfirm := widget.NewPasswordEntry()
//	registerConfirm.SetPlaceHolder("Подтверждение пароля")
//
//	statusLabel := widget.NewLabel("")
//	progress := widget.NewProgressBarInfinite()
//	progress.Hidden = true
//
//	loginTab := container.NewVBox(
//		widget.NewLabelWithStyle("Вход", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//		loginEmail,
//		loginPassword,
//		widget.NewButton("Войти", func() {
//			progress.Hidden = false
//			statusLabel.SetText("")
//
//			go func() {
//				tokens, err := login(loginEmail.Text, loginPassword.Text)
//				if err != nil {
//					fyne.Do(func() {
//						progress.Hidden = true
//						statusLabel.SetText("Ошибка: " + err.Error())
//					})
//					return
//				}
//
//				// Получаем пользователя
//				user, err := getCurrentUser(tokens.AccessToken)
//				if err != nil {
//					fyne.Do(func() {
//						progress.Hidden = true
//						statusLabel.SetText("Ошибка получения пользователя: " + err.Error())
//					})
//					return
//				}
//
//				currentUser = user
//				currentTokens = tokens
//				saveTokensToFile(tokens)
//				startTokenRefreshRoutine(myApp)
//
//				fyne.Do(func() {
//					parentWindow.Close()
//					authWin.Close()
//					showMainChatWindow(myApp, tokens)
//				})
//			}()
//		}),
//	)
//
//	registerTab := container.NewVBox(
//		widget.NewLabelWithStyle("Регистрация", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//		registerEmail,
//		registerUsername,
//		registerPassword,
//		registerConfirm,
//		widget.NewButton("Зарегистрироваться", func() {
//			if registerPassword.Text != registerConfirm.Text {
//				statusLabel.SetText("Пароли не совпадают")
//				return
//			}
//			if len(registerUsername.Text) < 5 {
//				statusLabel.SetText("Имя пользователя должно быть не менее 5 символов")
//				return
//			}
//
//			progress.Hidden = false
//			statusLabel.SetText("")
//
//			go func() {
//				tokens, err := register(registerEmail.Text, registerUsername.Text, registerPassword.Text)
//				if err != nil {
//					fyne.Do(func() {
//						progress.Hidden = true
//						statusLabel.SetText("Ошибка: " + err.Error())
//					})
//					return
//				}
//
//				// Получаем пользователя
//				user, err := getCurrentUser(tokens.AccessToken)
//				if err != nil {
//					fyne.Do(func() {
//						progress.Hidden = true
//						statusLabel.SetText("Ошибка получения пользователя: " + err.Error())
//					})
//					return
//				}
//
//				currentUser = user
//				currentTokens = tokens
//				saveTokensToFile(tokens)
//				startTokenRefreshRoutine(myApp)
//
//				fyne.Do(func() {
//					parentWindow.Close()
//					authWin.Close()
//					showMainChatWindow(myApp, tokens)
//				})
//			}()
//		}),
//	)
//
//	tabs := container.NewAppTabs(
//		container.NewTabItem("Вход", loginTab),
//		container.NewTabItem("Регистрация", registerTab),
//	)
//
//	content := container.NewVBox(
//		tabs,
//		widget.NewSeparator(),
//		progress,
//		statusLabel,
//	)
//
//	authWin.SetContent(content)
//	parentWindow.Hide()
//	authWin.Show()
//}
//
//// --- Вспомогательные функции ---
//
//func unique(ids []uint64) []uint64 {
//	seen := make(map[uint64]bool)
//	result := []uint64{}
//
//	for _, id := range ids {
//		if !seen[id] {
//			seen[id] = true
//			result = append(result, id)
//		}
//	}
//	return result
//}
