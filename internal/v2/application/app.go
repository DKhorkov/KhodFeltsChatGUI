package application

import (
	"context"

	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
)

// App struct - основной структура приложения Wails.
type App struct {
	useCases     interfaces.UseCases
	logger       logging.Logger
	errorsMapper interfaces.ErrorsMapper

	handlers []interfaces.Handler
}

// New создает новый экземпляр приложения.
func New(
	useCases interfaces.UseCases,
	logger logging.Logger,
	errorsMapper interfaces.ErrorsMapper,
	handlers []interfaces.Handler,
) *App {
	return &App{
		useCases:     useCases,
		logger:       logger,
		errorsMapper: errorsMapper,
		handlers:     handlers,
	}
}

// Startup вызывается при запуске приложения.
func (a *App) Startup(ctx context.Context) {
	// Инициализируем хендлеры с контекстом
	for _, handler := range a.handlers {
		handler.SetContext(ctx)
	}

	logging.LogInfo(a.logger, "Приложение успешно запущено")
}

// Shutdown вызывается при закрытии приложения.
func (a *App) Shutdown(_ context.Context) {
	// Останавливаем все горутины в хендлерах
	for _, handler := range a.handlers {
		handler.StopListening()
	}

	logging.LogInfo(a.logger, "Приложение успешно остановлено")
}

// BindHandlers Биндим хендлеры для доступа из фронтенда.
func (a *App) BindHandlers() []any {
	result := make([]any, 0, len(a.handlers))
	for _, handler := range a.handlers {
		result = append(result, handler)
	}

	return result
}
