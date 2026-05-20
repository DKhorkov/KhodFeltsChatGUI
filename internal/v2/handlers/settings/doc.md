# Пакет internal/v2/handlers/settings

## Назначение

Хендлер пользовательских настроек. Получение и обновление настроек через use cases.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер настроек. Содержит `useCases`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetSettings() (*domains.Settings, error)` | Возвращает текущие настройки пользователя. |
| `(h *Handler) UpdateSettings(settings domains.Settings) (*domains.Settings, error)` | Обновляет настройки и возвращает обновлённый результат. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `Settings`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`
