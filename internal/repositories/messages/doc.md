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
| `DeleteMessage(ctx, accessToken, dto)` | DELETE | `/messages/{id}` | Удаление сообщения. Body JSON: `{"forAll": bool}`. Ожидает 204 NoContent |

## Константы

- `accessTokenCookieName` = `"accessToken"`
- `limitQueryParamName` = `"limit"`
- `offsetQueryParamName` = `"offset"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `Message`, `Pagination`, `DeleteMessageDTO`
- `internal/common` — HTTP-заголовки (`ContentTypeHeaderName`, `ApplicationJSONContentType`)
