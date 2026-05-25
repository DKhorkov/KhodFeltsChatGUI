# Пакет internal/repositories/chats

## Назначение

HTTP-репозиторий для работы с чатами. Реализует интерфейс `interfaces.ChatsRepository`. Потокобезопасен: чтение через `RLock`, запись через `Lock` (`sync.RWMutex`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.RWMutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `GetUserChats(ctx, accessToken, pagination)` | GET | `/chats` | Список чатов пользователя с пагинацией (`limit`, `offset`) |
| `CreateChat(ctx, accessToken, chat)` | POST | `/chats` | Создание чата, возвращает `*domains.Chat` |

## Константы

- `accessTokenCookieName` = `"accessToken"`
- `limitQueryParamName` = `"limit"`
- `offsetQueryParamName` = `"offset"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `Chat`, `Pagination`
- `internal/common` — HTTP-заголовки
