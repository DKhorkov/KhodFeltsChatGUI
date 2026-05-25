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
	messagesrepository "github.com/DKhorkov/kfcGUI/internal/repositories/messages"
	settingsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/settings"
	tokensrepository "github.com/DKhorkov/kfcGUI/internal/repositories/tokens"
	usersrepository "github.com/DKhorkov/kfcGUI/internal/repositories/users"
	wsrepository "github.com/DKhorkov/kfcGUI/internal/repositories/ws"
	"github.com/DKhorkov/kfcGUI/internal/usecases"
	"github.com/DKhorkov/kfcGUI/internal/v2/application"
	authhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/auth"
	chatshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/chats"
	messageshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/messages"
	notificationhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/notifications"
	settingshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/settings"
	themehandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/theme"
	usershandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/users"
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
	tokensRepository := tokensrepository.New()
	settingsRepository := settingsrepository.New(httpClient, cfg.HTTP.BaseURL)
	websocketsRepository := wsrepository.New(cfg.HTTP.WebsocketURL, logger)

	errorsMapper := errors.New()

	useCases := usecases.New(
		usersRepository,
		chatsRepository,
		messagesRepository,
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

	chatsHandler := chatshandler.New(
		useCases,
		errorsMapper,
		logger,
	)

	messagesHandler := messageshandler.New(useCases)

	usersHandler := usershandler.New(
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
			chatsHandler,
			messagesHandler,
			usersHandler,
			themeHandler,
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
