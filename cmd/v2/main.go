package main

import (
	"context"
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
	authhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/auth"
	chathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/chat"
	createchathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/create_chat"
	forgetpasswordhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/forget_password"
	searchusershandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/search_users"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct - основной структура приложения Wails
type App struct {
	ctx context.Context

	// Хендлеры
	authHandler           *authhandler.Handler
	chatHandler           *chathandler.Handler
	createChatHandler     *createchathandler.Handler
	searchUsersHandler    *searchusershandler.Handler
	forgetPasswordHandler *forgetpasswordhandler.Handler

	// Сервисы
	useCases     interfaces.UseCases
	logger       logging.Logger
	errorsMapper interfaces.ErrorsMapper
}

// NewApp создает новый экземпляр приложения
func NewApp(
	authHandler *authhandler.Handler,
	chatHandler *chathandler.Handler,
	createChatHandler *createchathandler.Handler,
	searchUsersHandler *searchusershandler.Handler,
	forgetPasswordHandler *forgetpasswordhandler.Handler,
	useCases interfaces.UseCases,
	logger logging.Logger,
	errorsMapper interfaces.ErrorsMapper,
) *App {
	return &App{
		authHandler:           authHandler,
		chatHandler:           chatHandler,
		createChatHandler:     createChatHandler,
		searchUsersHandler:    searchUsersHandler,
		forgetPasswordHandler: forgetPasswordHandler,
		useCases:              useCases,
		logger:                logger,
		errorsMapper:          errorsMapper,
	}
}

// Startup вызывается при запуске приложения
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Инициализируем хендлеры с контекстом
	if startupHandler, ok := any(a.chatHandler).(interface{ Startup(context.Context) }); ok {
		startupHandler.Startup(ctx)
	}

	logging.LogInfo(a.logger, "Приложение успешно запущено")
}

// Shutdown вызывается при закрытии приложения
func (a *App) Shutdown(ctx context.Context) {
	// Останавливаем все горутины в хендлерах
	if a.chatHandler != nil {
		a.chatHandler.StopListening()
	}

	logging.LogInfo(a.logger, "Приложение успешно остановлено")
}

// Биндим хендлеры для доступа из фронтенда
func (a *App) BindHandlers() []any {
	return []any{
		a.authHandler,
		a.chatHandler,
		a.createChatHandler,
		a.searchUsersHandler,
		a.forgetPasswordHandler,
	}
}

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

	authRepository := authrepository.New(httpClient, cfg.HTTP.BaseURL)
	usersRepository := usersrepository.New(httpClient, cfg.HTTP.BaseURL)
	chatsRepository := chatsrepository.New(httpClient, cfg.HTTP.BaseURL)
	tokensRepository := tokensrepository.New()
	settingsRepository := settingsrepository.New()
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
		cfg.Validation,
	)

	createChatHandler := createchathandler.New(useCases)

	searchUsersHandler := searchusershandler.New(useCases)

	forgetPasswordHandler := forgetpasswordhandler.New(
		useCases,
		errorsMapper,
		cfg.Validation,
	)

	// 8. Создаем главное приложение
	app := NewApp(
		authHandler,
		chatHandler,
		createChatHandler,
		searchUsersHandler,
		forgetPasswordHandler,
		useCases,
		logger,
		errorsMapper,
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
