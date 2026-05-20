# Пакет internal/repositories/tokens

## Назначение

Локальное хранилище токенов авторизации в файловой системе. Реализует интерфейс `interfaces.TokensRepository`. Токены сохраняются в JSON-файл `tokens.json` в директории данных приложения (`common.AppDataDir()`). Потокобезопасен (`sync.RWMutex`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Содержит `mu sync.RWMutex` |

## Методы

| Метод | Описание |
|-------|----------|
| `Save(ctx, tokens)` | Сериализует `domains.TokensDTO` в JSON и записывает в файл (права `0o600`) |
| `Load(ctx)` | Читает и десериализует токены из файла |
| `Delete(ctx)` | Удаляет файл токенов |

## Константы

- `tokensFilename` = `"tokens.json"`
- `permission` = `0o600`
- Путь к файлу: `common.AppDataDir() + "/tokens.json"`

## Зависимости

- `internal/common` — `AppDataDir()` для определения пути к файлу
- `internal/domains` — `TokensDTO`
