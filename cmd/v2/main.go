package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/config"
	"github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	authrepository "github.com/DKhorkov/kfcGUI/internal/repositories/auth"
	chatsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/chats"
	settingsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/settings"
	tokensrepository "github.com/DKhorkov/kfcGUI/internal/repositories/tokens"
	usersrepository "github.com/DKhorkov/kfcGUI/internal/repositories/users"
	wsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/ws"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	"github.com/DKhorkov/kfcGUI/internal/v2/application"
	authhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/auth"
	chathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/chat"
	createchathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/create_chat"
	forgetpasswordhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/forget_password"
	notificationhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/notification"
	profilehandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/profile"
	searchusershandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/search_users"
	settingshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/settings"
	themehandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/theme"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Инициализируем переменные окружения для дальнейшего считывания в конфиге:
	loadenv.Init()

	common.CreateAppDataDir()
	common.CreateLogsDir()
	common.CopyAppIcon()

	cfg := config.New()

	logger := logging.New(
		cfg.Logging.Level,
		cfg.Logging.LogFilePath,
	)

	httpClient := &http.Client{Timeout: cfg.HTTP.Timeout}

	authRepository := authrepository.New(httpClient, cfg.HTTP.BaseURL)
	usersRepository := usersrepository.New(httpClient, cfg.HTTP.BaseURL)
	chatsRepository := chatsrepository.New(httpClient, cfg.HTTP.BaseURL)
	tokensRepository := tokensrepository.New()
	settingsRepository := settingsrepository.New(httpClient, cfg.HTTP.BaseURL)
	websocketsRepository := wsrepository.New(cfg.HTTP.WebsocketURL, logger)

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

	// 7. Создаем хендлеры
	authHandler := authhandler.New(
		useCases,
		errorsMapper,
		cfg.Validation,
	)

	chatHandler := chathandler.New(
		useCases,
		errorsMapper,
		logger,
		cfg.Validation,
	)

	createChatHandler := createchathandler.New(useCases, errorsMapper)

	searchUsersHandler := searchusershandler.New(useCases)

	forgetPasswordHandler := forgetpasswordhandler.New(
		useCases,
		errorsMapper,
		cfg.Validation,
	)

	profileHandler := profilehandler.New(
		useCases,
		errorsMapper,
		cfg.Validation,
	)

	themeHandler := themehandler.New(useCases)
	settingsHandler := settingshandler.New(useCases)
	notificationHandler := notificationhandler.New(logger)

	// 8. Создаем главное приложение
	app := application.New(
		useCases,
		logger,
		errorsMapper,
		[]interfaces.Handler{
			authHandler,
			chatHandler,
			createChatHandler,
			searchUsersHandler,
			forgetPasswordHandler,
			themeHandler,
			profileHandler,
			settingsHandler,
			notificationHandler,
		},
	)

	// 9. Запускаем Wails приложение
	err := wails.Run(&options.App{
		Title:     "KFC Chat",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind:             app.BindHandlers(),
	})
	if err != nil {
		logging.LogError(logger, "Ошибка запуска приложения", err)
		log.Fatal(err)
	}
}
