# Пакет internal/repositories/reactions

## Назначение

HTTP-репозиторий для работы с реакциями на сообщения. Реализует интерфейс `interfaces.ReactionsRepository`. Потокобезопасен: чтение через `RLock`, запись через `Lock` (`sync.RWMutex`).

Выделен из `repositories/messages` для соблюдения принципа единственной ответственности: реакции — отдельная сущность со своим справочником и M2M-таблицей на бэкенде.

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.RWMutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `ListReactions(ctx)` | GET | `/reactions` | Публичный справочник emoji для UI-пикера. Cookie с `accessToken` не отправляется — роут исключён из auth-middleware на сервере. Ожидает 200 OK. Read lock |
| `AddMessageReaction(ctx, accessToken, dto)` | POST | `/messages/{id}/reactions` | Body JSON `{"reactionId": N}`. Ожидает 204 NoContent. Маппинг ошибок: 409 → `errors.ErrReactionAlreadyExists`, 404 → `errors.ErrReactionNotFound`. Write lock |
| `RemoveMessageReaction(ctx, accessToken, dto)` | DELETE | `/messages/{id}/reactions/{reactionId}` | Идемпотентно: 204 всегда при успехе (сервер сам решает, публиковать ли WS). Write lock |

## Константы

- `accessTokenCookieName` = `"accessToken"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `Reaction`, `MessageReactionDTO`
- `internal/errors` — sentinel-ошибки `ErrReactionAlreadyExists`, `ErrReactionNotFound`
- `internal/common` — HTTP-заголовки (`ContentTypeHeaderName`, `ApplicationJSONContentType`)
