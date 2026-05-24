# Архитектура бэкенда

## Паттерн

Бэкенд построен по трёхслойной архитектуре:

```
Handler -> UseCases -> Repositories
```

Каждый слой взаимодействует с соседним через Go-интерфейсы, объявленные в `internal/interfaces/`. Это обеспечивает возможность подмены реализаций и упрощает тестирование с использованием моков.

## Точка входа

Файл: `cmd/v2/main.go`

`main()` выполняет следующие шаги по порядку:

1. Загружает переменные окружения (`loadenv.Init()`).
2. Создаёт директории для данных приложения и логов.
3. Инициализирует конфигурацию (`config.New()`).
4. Создаёт логгер.
5. Создаёт HTTP-клиент с таймаутом из конфига.
6. Создаёт все репозитории (auth, users, chats, tokens, settings, ws).
7. Создаёт маппер ошибок (`errors.New()`).
8. Создаёт UseCase-слой (`usecases.New(...)`), передавая ему все репозитории.
9. Создаёт хендлеры, передавая каждому UseCases и ErrorsMapper.
10. Создаёт `App` и запускает `wails.Run(...)`.

## Жизненный цикл приложения

Файл: `internal/v2/application/app.go`

Структура `App` содержит:
- `useCases interfaces.UseCases`
- `logger logging.Logger`
- `errorsMapper interfaces.ErrorsMapper`
- `handlers []interfaces.Handler`

### Startup
Вызывается Wails при старте приложения. Передаёт Wails-контекст каждому хендлеру через `handler.SetContext(ctx)` и логирует успешный запуск.

### Shutdown
Вызывается Wails при закрытии приложения. Вызывает `handler.StopListening()` у каждого хендлера, дожидается завершения фоновых горутин и логирует остановку.

### BindHandlers
Возвращает срез `[]any` со всеми хендлерами. Wails биндит эти объекты, автоматически генерируя TypeScript-обёртки для каждого публичного метода.

```go
wails.Run(&options.App{
    OnStartup:  app.Startup,
    OnShutdown: app.Shutdown,
    Bind:       app.BindHandlers(),
})
```

Биндятся следующие хендлеры:
- `authHandler`
- `chatsHandler`
- `messagesHandler`
- `usersHandler`
- `themeHandler`
- `settingsHandler`
- `notificationsHandler`

## Параметры окна

Задаются непосредственно в `cmd/v2/main.go`:
- Заголовок: `"KFC Chat"`
- Размер: 1200x800 (минимум 800x600)
- Фронтенд встроен в бинарь через `//go:embed all:frontend/dist`

## Ключевые зависимости (`go.mod`)

| Пакет | Версия | Назначение |
|---|---|---|
| `github.com/wailsapp/wails/v2` | v2.12.0 | Десктопный фреймворк (Go + WebView) |
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket-соединение для реального времени |
| `github.com/DKhorkov/libs` | v1.14.10 | Логирование, валидация, загрузка env |
| `fyne.io/fyne/v2` | v2.7.3 | Устаревший UI-фреймворк (v1) |
| `github.com/golang-jwt/jwt/v5` | v5.2.1 | JWT-токены |

## Легаси v1 (Fyne)

Точка входа: `cmd/v1/main.go`. Код v1 находится в `internal/v1/`. Использует фреймворк Fyne для построения нативного UI. Не является активным путём — приложение запускается через `cmd/v2/main.go`.
