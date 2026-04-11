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
	"github.com/DKhorkov/kfcGUI/internal/repositories/auth"
	"github.com/DKhorkov/kfcGUI/internal/repositories/chats"
	"github.com/DKhorkov/kfcGUI/internal/repositories/settings"
	"github.com/DKhorkov/kfcGUI/internal/repositories/tokens"
	"github.com/DKhorkov/kfcGUI/internal/repositories/users"
	"github.com/DKhorkov/kfcGUI/internal/repositories/ws"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	authwindow "github.com/DKhorkov/kfcGUI/internal/windows/auth"
	"github.com/DKhorkov/kfcGUI/internal/windows/chat"
	createChat "github.com/DKhorkov/kfcGUI/internal/windows/create_chat"
	forgetPassword "github.com/DKhorkov/kfcGUI/internal/windows/forget_password"
	"github.com/DKhorkov/kfcGUI/internal/windows/information"
	"github.com/DKhorkov/kfcGUI/internal/windows/notification"
	searchUsers "github.com/DKhorkov/kfcGUI/internal/windows/search_users"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
)

var passwordChangedInformationWindowTitle = "Пароль успешно сброшен" //nolint:gosec // наименование переменной

func main() {
	// Инициализируем переменные окружения для дальнейшего считывания в конфиге:
	loadenv.Init()

	common.CreateLogsDir()

	cfg := config.New()

	logger := logging.New(
		cfg.Logging.Level,
		cfg.Logging.LogFilePath,
	)

	httpClient := &http.Client{Timeout: cfg.HTTP.Timeout}

	authRepository := auth.New(httpClient, cfg.HTTP.BaseURL)
	usersRepository := users.New(httpClient, cfg.HTTP.BaseURL)
	chatsRepository := chats.New(httpClient, cfg.HTTP.BaseURL)
	tokensRepository := tokens.New()
	settingsRepository := settings.New()
	websocketsRepository := ws.New(cfg.HTTP.WebsocketURL, logger)

	errorsMapper := errors.New()

	useCases := usecases.New(
		usersRepository,
		chatsRepository,
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

	forgetPasswordWindow := forgetPassword.New(
		kfc,
		information.New(kfc, passwordChangedInformationWindowTitle),
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
	notificationWindow := notification.New(kfc, useCases)
	searchUsersWindow := searchUsers.New(kfc, useCases)
	createChatWindow := createChat.New(kfc, useCases, nil)
	chatWindow := chat.New(
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
