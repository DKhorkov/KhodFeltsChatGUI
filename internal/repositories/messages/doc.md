# Пакет internal/repositories/messages

## Назначение

HTTP-репозиторий для работы с сообщениями. Реализует интерфейс `interfaces.MessagesRepository`. Потокобезопасен: чтение через `RLock`, запись через `Lock` (`sync.RWMutex`).

Выделен из `repositories/chats` — методы `GetChatMessages` и `DeleteMessage` перенесены в отдельный пакет для соблюдения принципа единственной ответственности.

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.RWMutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `GetChatMessages(ctx, accessToken, chatID, pagination)` | GET | `/chats/{id}/messages` | Сообщения конкретного чата с пагинацией (`limit`, `offset`) |
| `GetMessageByID(ctx, accessToken, messageID)` | GET | `/messages/{id}` | Получение одного сообщения по ID. Ожидает 200 OK |
| `DeleteMessage(ctx, accessToken, dto)` | DELETE | `/messages/{id}` | Удаление сообщения. Body JSON: `{"forAll": bool}`. Ожидает 204 NoContent |
| `UpdateMessage(ctx, accessToken, dto)` | PUT | `/messages/{id}` | Редактирование сообщения. Body JSON с новым текстом. Ожидает 204 NoContent. Использует write lock (`r.mu.Lock()`) |
| `ListReactions(ctx, accessToken)` | GET | `/reactions` | Справочник emoji для UI-пикера. Ожидает 200 OK. Read lock |
| `AddMessageReaction(ctx, accessToken, dto)` | POST | `/messages/{id}/reactions` | Body JSON `{"reactionId": N}`. Ожидает 204 NoContent. Ошибка тела на 409 — «уже стоит». Write lock |
| `RemoveMessageReaction(ctx, accessToken, dto)` | DELETE | `/messages/{id}/reactions/{reactionId}` | Идемпотентно: 204 всегда при успехе (сервер сам решает, публиковать ли WS). Write lock |

## Константы

- `accessTokenCookieName` = `"accessToken"`
- `limitQueryParamName` = `"limit"`
- `offsetQueryParamName` = `"offset"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `Message`, `Pagination`, `DeleteMessageDTO`, `UpdateMessageDTO`, `Reaction`, `MessageReactionDTO`
- `internal/common` — HTTP-заголовки (`ContentTypeHeaderName`, `ApplicationJSONContentType`)
