# frontend/src/styles

## Назначение

Глобальные CSS-стили и дизайн-токены приложения. Файл `global.css` подключается на уровне `index.html`.

## Дизайн-токены (CSS-переменные)

Определены в `:root` (светлая тема) и переопределены в `[data-bs-theme="dark"]` (тёмная тема).

### Группы переменных

| Группа | Переменные | Пример |
|--------|-----------|--------|
| Радиусы | `--radius-sm`, `--radius-md`, `--radius-lg`, `--radius-full` | `6px`, `8px`, `12px`, `50%` |
| Тени | `--shadow-sm`, `--shadow-md`, `--shadow-lg` | `0 2px 4px rgba(...)` |
| Отступы | `--space-xs` .. `--space-3xl` | `4px` .. `32px` |
| Шрифты | `--font-xs` .. `--font-xl` | `12px` .. `18px` |
| Анимации | `--transition-fast`, `--transition-base`, `--transition-slow` | `0.15s` .. `0.3s` |
| Фоны | `--bg-app`, `--bg-panel`, `--bg-surface`, `--bg-hover`, `--bg-active`, `--bg-input`, `--bg-message-other`, `--bg-danger`, `--bg-info`, `--bg-toast` | — |
| Границы | `--border`, `--border-input`, `--border-danger`, `--border-danger-hover` | — |
| Текст | `--text-primary`, `--text-secondary`, `--text-muted`, `--text-placeholder`, `--text-on-accent`, `--text-toast` | — |
| Акценты | `--accent`, `--accent-hover`, `--accent-secondary`, `--danger`, `--info`, `--info-text` | — |

## Темы

Переключение тем реализовано через атрибут `data-bs-theme` на `<html>`:
- `light` (по умолчанию) — `:root`
- `dark` — `[data-bs-theme="dark"]`

## Общие CSS-классы

В `global.css` также определены стили для:
- Модальных окон (`.modal-overlay`, `.modal-content`)
- Кнопок (`.btn-primary`, `.btn-danger`, `.btn-close`)
- Форм (`.form-group`, `.form-input`, `.form-label`)
- Профильных модалок (`.profile-modal`)
- Контекстного меню аватара (`.avatar-context-menu`, `.avatar-context-menu__item`)
- Кликабельного аватара и оверлея увеличения (`.profile-modal__avatar--clickable`, `.avatar-zoom-overlay`) — общий для ProfileModal и ChatView
- Уведомлений (`.notification-toast`)
- Загрузочного экрана (`.loading-screen`)
