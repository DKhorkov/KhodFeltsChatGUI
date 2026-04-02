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
	"github.com/DKhorkov/kfcGUI/internal/repositories"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	"github.com/DKhorkov/kfcGUI/internal/windows/auth"
	"github.com/DKhorkov/kfcGUI/internal/windows/chat"
	createChat "github.com/DKhorkov/kfcGUI/internal/windows/create_chat"
	forgetPassword "github.com/DKhorkov/kfcGUI/internal/windows/forget_password"
	"github.com/DKhorkov/kfcGUI/internal/windows/information"
	"github.com/DKhorkov/kfcGUI/internal/windows/notification"
	searchUsers "github.com/DKhorkov/kfcGUI/internal/windows/search_users"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
)

var passwordChangedInformationWindowTitle = "Пароль успешно сброшен"

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

	authRepository := repositories.NewAuthRepository(httpClient, cfg.HTTP.BaseURL)
	usersRepository := repositories.NewUsersRepository(httpClient, cfg.HTTP.BaseURL)
	chatsRepository := repositories.NewChatsRepository(httpClient, cfg.HTTP.BaseURL)
	tokensRepository := repositories.NewTokensRepository()
	settingsRepository := repositories.NewSettingsRepository()
	websocketsRepository := repositories.NewWebSocketsRepository(cfg.HTTP.WebsocketURL, logger)

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
	authWindow := auth.New(kfc, nil, forgetPasswordWindow, useCases, cfg.Validation, errorsMapper)
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
