package main

import (
	"context"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	authrepository "github.com/DKhorkov/kfcGUI/internal/repositories/auth"
	chatsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/chats"
	messagesrepository "github.com/DKhorkov/kfcGUI/internal/repositories/messages"
	reactionsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/reactions"
	settingsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/settings"
	tokensrepository "github.com/DKhorkov/kfcGUI/internal/repositories/tokens"
	usersrepository "github.com/DKhorkov/kfcGUI/internal/repositories/users"
	wsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/ws"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	authwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/auth"
	chatwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/chat"
	createchatwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/create_chat"
	forgetpasswordwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/forget_password"
	informationwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/information"
	notificationwindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/notification"
	searchuserswindow "github.com/DKhorkov/kfcGUI/internal/v1/windows/search_users"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
)

var passwordChangedInformationWindowTitle = "Пароль успешно сброшен" //nolint:gosec // наименование переменной

func main() {
	// Инициализируем переменные окружения для дальнейшего считывания в конфиге:
	loadenv.Init()

	common.CreateAppDataDir()
	common.CreateLogsDir()

	cfg := config.New()

	logger := logging.New(
		cfg.Logging.Level,
		cfg.Logging.LogFilePath,
	)

	httpClient := &http.Client{Timeout: cfg.HTTP.Timeout}

	authRepository := authrepository.New(httpClient, cfg.HTTP.BaseURL)
	usersRepository := usersrepository.New(httpClient, cfg.HTTP.BaseURL)
	chatsRepository := chatsrepository.New(httpClient, cfg.HTTP.BaseURL)
	messagesRepository := messagesrepository.New(httpClient, cfg.HTTP.BaseURL)
	reactionsRepository := reactionsrepository.New(httpClient, cfg.HTTP.BaseURL)
	tokensRepository := tokensrepository.New()
	settingsRepository := settingsrepository.New(httpClient, cfg.HTTP.BaseURL)
	websocketsRepository := wsrepository.New(cfg.HTTP.WebsocketURL, logger)

	errorsMapper := errors.New()

	useCases := usecases.New(
		usersRepository,
		chatsRepository,
		messagesRepository,
		reactionsRepository,
		authRepository,
		tokensRepository,
		settingsRepository,
		websocketsRepository,
		logger,
		errorsMapper,
	)

	appTheme := theme.LightTheme()
	if useCases.GetTheme(context.Background()) == domains.ThemeDark {
		appTheme = theme.DarkTheme()
	}

	kfc := app.New()
	kfc.Settings().SetTheme(appTheme)

	forgetPasswordWindow := forgetpasswordwindow.New(
		kfc,
		informationwindow.New(kfc, passwordChangedInformationWindowTitle),
		useCases,
		cfg.Validation,
		errorsMapper,
	)
	authWindow := authwindow.New(
		kfc,
		nil,
		forgetPasswordWindow,
		useCases,
		cfg.Validation,
		errorsMapper,
	)
	notificationWindow := notificationwindow.New(kfc, useCases)
	searchUsersWindow := searchuserswindow.New(kfc, useCases)
	createChatWindow := createchatwindow.New(kfc, useCases, nil)
	chatWindow := chatwindow.New(
		kfc,
		authWindow,
		createChatWindow,
		searchUsersWindow,
		notificationWindow,
		logger,
		useCases,
	)

	authWindow.SetChatWindow(chatWindow)
	createChatWindow.SetRefreshChatsFunc(chatWindow.RefreshChats)

	go func() {
		fyne.Do(func() {
			if _, err := useCases.Authenticate(context.Background()); err != nil {
				authWindow.Build(nil)
				authWindow.Show()

				return
			}

			chatWindow.Build(nil)
			chatWindow.Show()
		})
	}()

	kfc.Run()
}
