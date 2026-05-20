# Пакет internal/repositories/base

## Назначение

Базовый репозиторий, содержащий общую функциональность для всех HTTP-репозиториев. Предоставляет безопасное закрытие тела HTTP-ответа с логированием ошибок. Встраивается в конкретные репозитории через композицию (`base.Repository`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Базовая структура с полем `logger logging.Logger` |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(logger logging.Logger) *Repository` | Конструктор |
| `(r *Repository) CloseBody(ctx context.Context, body io.ReadCloser)` | Закрывает тело ответа; при ошибке логирует через `logging.LogErrorContext` |

## Зависимости

- `github.com/DKhorkov/libs/logging` — логирование
