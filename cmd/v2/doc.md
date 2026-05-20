# Пакет cmd/v2

## Назначение

Точка входа Wails v2 приложения. Инициализирует всю инфраструктуру и запускает десктопное окно.

## Порядок инициализации в main.go

1. Загружаются переменные окружения (`loadenv.Init`), создаются директории `AppData` и `Logs`.
2. Читается конфиг (`config.New()`), создаётся логгер.
3. Создаётся `http.Client` с таймаутом из конфига.
4. Инициализируются репозитории:
   - `auth`, `users`, `chats` — HTTP-репозитории (используют `httpClient` + `baseURL`).
   - `tokens` — файловое хранилище токенов.
   - `settings` — HTTP-репозиторий настроек.
   - `ws` — WebSocket-репозиторий (используют `websocketURL`).
5. Создаётся маппер ошибок (`errors.New()`).
6. Создаётся единый `UseCases` со всеми репозиториями.
7. Создаются 9 хендлеров: `auth`, `chat`, `create_chat`, `search_users`, `forget_password`, `profile`, `theme`, `settings`, `notification`.
8. Создаётся `application.App` со списком хендлеров.
9. Запускается Wails-приложение с параметрами окна (1200×800, min 800×600).

## Embed

`//go:embed all:frontend/dist` — статические ассеты фронтенда встраиваются в бинарник.

## Зависимости

- `github.com/wailsapp/wails/v2` — фреймворк десктопного приложения.
- `github.com/DKhorkov/libs/loadenv`, `logging` — утилиты.
- Все внутренние пакеты `internal/`.
