# Пакет internal/v2/handlers/theme

## Назначение

Хендлер управления темой оформления. Получение текущей темы и переключение между светлой и тёмной.

## Типы

| Тип | Описание |
|-----|----------|
| `Handler` | Wails-хендлер темы. Содержит `useCases`, `wailsCtx`. |

## Функции и методы

| Сигнатура | Описание |
|-----------|----------|
| `New(useCases interfaces.UseCases) *Handler` | Конструктор хендлера. |
| `(h *Handler) SetContext(ctx context.Context)` | Устанавливает Wails-контекст. |
| `(h *Handler) GetTheme() domains.ThemeType` | Возвращает текущую тему. |
| `(h *Handler) ToggleTheme() (domains.ThemeType, error)` | Переключает тему: светлая -> тёмная, тёмная -> светлая. Возвращает новую тему. |
| `(h *Handler) StartListening()` | Заглушка (не реализовано). |
| `(h *Handler) StopListening()` | Заглушка (не реализовано). |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/domains` — `ThemeType`, `ThemeLight`, `ThemeDark`
- `github.com/DKhorkov/kfcGUI/internal/interfaces` — `UseCases`
