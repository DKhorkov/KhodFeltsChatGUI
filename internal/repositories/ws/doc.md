# Пакет internal/repositories/ws

## Назначение

WebSocket-репозиторий для обмена сообщениями в реальном времени. Реализует интерфейс `interfaces.WebSocketsRepository`. Использует библиотеку `gorilla/websocket`. Потокобезопасен (`sync.Mutex`).

Архитектура чтения: при подключении запускается горутина `readLoop`, которая непрерывно читает из сокета WS-события (`WSEvent` envelope) и складывает их в буферизированный канал (`eventsChan`, размер 100). Ошибки чтения передаются через отдельный канал (`errChan`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Содержит `baseURL`, `logger`, `ws *websocket.Conn`, `mu sync.Mutex`, `eventsChan`, `errChan` |

## Методы

| Метод | Описание |
|-------|----------|
| `Connect(ctx, accessToken)` | Устанавливает WebSocket-соединение на `/ws` с access-токеном в cookie; запускает `readLoop` |
| `Close()` | Закрывает соединение; идемпотентен (безопасен при повторном вызове) |
| `ReadEvent(ctx)` | Читает WS-событие (`WSEvent`) из канала; поддерживает отмену через `ctx`; возвращает ошибку при закрытии соединения |
| `WriteMessage(ctx, message)` | Отправляет сообщение в JSON-формате с дедлайном 2 секунды; обрабатывает close-ошибки сокета |

## Константы

- `readMessagesBufferSize` = `100`
- `readErrorsBufferSize` = `1`
- `writeDeadline` = `2s`
- `accessTokenCookieName` = `"accessToken"`

## Зависимости

- `github.com/gorilla/websocket` — WebSocket-клиент
- `github.com/DKhorkov/libs/logging` — логирование
- `internal/domains` — `Message`
- `internal/errors` — `ErrWebsocket`, `ErrWebsocketClosed`
- `internal/common` — `CookieHeaderName`
