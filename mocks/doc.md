# mocks — Сгенерированные моки

## Назначение

Моки всех интерфейсов, сгенерированные через `go.uber.org/mock/mockgen`.

## Содержимое

| Директория | Моки для |
|-----------|----------|
| `application/` | Application (Startup/Shutdown/BindHandlers) |
| `errors/` | ErrorsMapper |
| `handler/` | Handler |
| `http/` | HTTPClient |
| `repositories/` | AuthRepository, ChatsRepository, SettingsRepository, TokensRepository, UsersRepository, WebsocketsRepository |
| `usecases/` | UseCases |
| `window/` | Window (v1) |

## Генерация

Моки генерируются на основе `//go:generate mockgen` директив в `internal/interfaces/`.
