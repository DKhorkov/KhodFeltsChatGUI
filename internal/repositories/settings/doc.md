# Пакет internal/repositories/settings

## Назначение

HTTP-репозиторий для работы с пользовательскими настройками. Реализует интерфейс `interfaces.SettingsRepository`. Потокобезопасен: чтение через `RLock`, запись через `Lock` (`sync.RWMutex`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.RWMutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `GetSettings(ctx, accessToken)` | GET | `/users/me/settings` | Получение настроек текущего пользователя |
| `UpdateSettings(ctx, accessToken, settings)` | PUT | `/users/me/settings` | Обновление настроек, возвращает `*domains.Settings` |

## Константы

- `accessTokenCookieName` = `"accessToken"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `Settings`
- `internal/common` — HTTP-заголовки
