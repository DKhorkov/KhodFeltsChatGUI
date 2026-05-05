# Стили фронтенда

Файл: `frontend/src/styles/global.css`

---

## Система дизайн-токенов (CSS custom properties)

Все визуальные константы определены как CSS-переменные в `:root`. Это позволяет переключать тему оформления только заменой значений переменных — без изменения компонентных стилей.

### Радиусы

| Переменная | Значение | Применение |
|---|---|---|
| `--radius-sm` | `6px` | Скроллбары, мелкие элементы |
| `--radius-md` | `8px` | Кнопки, поля ввода, большинство блоков |
| `--radius-lg` | `12px` | Модальные окна (`.modal-content`) |
| `--radius-full` | `50%` | Аватары (круглая форма) |

### Тени

| Переменная | Светлая тема | Тёмная тема |
|---|---|---|
| `--shadow-sm` | `0 2px 4px rgba(0,0,0,0.08)` | `0 2px 4px rgba(0,0,0,0.2)` |
| `--shadow-md` | `0 4px 12px rgba(0,0,0,0.12)` | `0 4px 12px rgba(0,0,0,0.3)` |
| `--shadow-lg` | `0 8px 24px rgba(0,0,0,0.16)` | `0 8px 24px rgba(0,0,0,0.4)` |

### Отступы

| Переменная | Значение |
|---|---|
| `--space-xs` | `4px` |
| `--space-sm` | `8px` |
| `--space-md` | `12px` |
| `--space-lg` | `16px` |
| `--space-xl` | `20px` |
| `--space-2xl` | `24px` |
| `--space-3xl` | `32px` |

### Типографика

| Переменная | Значение |
|---|---|
| `--font-xs` | `12px` |
| `--font-sm` | `13px` |
| `--font-md` | `14px` |
| `--font-lg` | `16px` |
| `--font-xl` | `18px` |

### Анимации

| Переменная | Значение | Применение |
|---|---|---|
| `--transition-fast` | `0.15s ease` | Появление модалок (`fadeIn`) |
| `--transition-base` | `0.2s ease` | Ховер-эффекты кнопок, полей, элементов списков |
| `--transition-slow` | `0.3s ease` | Переключатель темы |

---

## Поддержка тем (светлая / тёмная)

Тема выбирается через атрибут на корневом элементе HTML:

```html
<html data-bs-theme="dark">   <!-- тёмная тема -->
<html data-bs-theme="light">  <!-- светлая тема -->
```

Установка атрибута происходит в `App.js` при загрузке и в `ChatView.js` / `ProfileModal.js` при переключении через `toggleTheme()`.

Переменные фонов, границ и текста переопределяются в блоке `[data-bs-theme="dark"]`:

### Фоны

| Переменная | Светлая тема | Тёмная тема | Описание |
|---|---|---|---|
| `--bg-app` | `#ffffff` | `#1a1a2e` | Фон всего приложения |
| `--bg-panel` | `#f7f7f7` | `#16162a` | Боковые панели, секции профиля |
| `--bg-surface` | `#ffffff` | `#1e1e30` | Фон карточек и модалок |
| `--bg-hover` | `#f0f0f0` | `#252540` | Фон при наведении |
| `--bg-active` | `#e3e8ff` | `#2d3561` | Активный элемент списка |
| `--bg-input` | `#ffffff` | `#252540` | Фон полей ввода |
| `--bg-message-other` | `#f0f0f0` | `#252540` | Пузырь чужого сообщения |
| `--bg-danger` | `#fff5f5` | `#2d1515` | Фон элементов с ошибкой |
| `--bg-info` | `#e3f2fd` | `#0d2137` | Фон информационных блоков |
| `--bg-toast` | `#2a2a3e` | `#2e2e50` | Фон всплывающих уведомлений |

### Границы

| Переменная | Светлая тема | Тёмная тема |
|---|---|---|
| `--border` | `#e0e0e0` | `#2e2e50` |
| `--border-input` | `#dddddd` | `#3a3a60` |
| `--border-danger` | `#ffeeee` | `#4a2020` |
| `--border-danger-hover` | `#feb2b2` | `#6b3030` |

### Текст

| Переменная | Светлая тема | Тёмная тема | Применение |
|---|---|---|---|
| `--text-primary` | `#333333` | `#e2e2f0` | Основной текст |
| `--text-secondary` | `#555555` | `#b8b8d0` | Метки форм, вторичный текст |
| `--text-muted` | `#666666` | `#8888a8` | Email в профиле, дата |
| `--text-placeholder` | `#999999` | `#66668a` | Плейсхолдеры, скроллбар |
| `--text-on-accent` | `#ffffff` | `#ffffff` | Текст на акцентном фоне (кнопки) |
| `--text-toast` | `#e2e2f0` | `#e2e2f0` | Текст уведомлений |

### Акценты

| Переменная | Светлая тема | Тёмная тема | Описание |
|---|---|---|---|
| `--accent` | `#667eea` | `#667eea` | Основной акцент (кнопки, аватары, фокус) |
| `--accent-hover` | `#5a67d8` | `#7a8ef8` | Акцент при наведении |
| `--accent-secondary` | `#764ba2` | — | Только в светлой теме |
| `--danger` | `#e53e3e` | `#fc6b6b` | Цвет ошибок и опасных действий |
| `--info` | `#1976d2` | `#64b5f6` | Информационный акцент |
| `--info-text` | `#1976d2` | `#64b5f6` | Цвет текста информационных блоков |

---

## Базовые стили

```css
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, ...;
    font-size: var(--font-md);
    background-color: var(--bg-app);
    color: var(--text-primary);
}
```

Скроллбары кастомизированы для WebKit (ширина 6px, прозрачный трек, thumb из `--text-placeholder`).

---

## Общие стили компонентов

### Модальные окна

`.modal-overlay` — полноэкранный затемнённый фон (z-index: 1000). Появляется с анимацией `fadeIn var(--transition-fast)`. Директива `v-focus` захватывает фокус для обработки клавиши Escape.

`.modal-content` — карточка модалки:
- Ширина: `480px`, максимальная высота: `80vh` с прокруткой
- Фон: `var(--bg-surface)`, тень: `var(--shadow-lg)`, скругление: `var(--radius-lg)`
- Отступ: `var(--space-3xl)`

Дополнительные классы внутри `.modal-content`:
- `.modal-content__close` — кнопка закрытия (абсолютно позиционированная, `×`)
- `.modal-content__title` — заголовок (`font-size: var(--font-xl)`, `font-weight: 600`)
- `.modal-content__info` — информационный блок (фон `--bg-info`, цвет `--info-text`)
- `.modal-content__form-group` — обёртка поля с меткой
- `.modal-content__users-list` — прокручиваемый список пользователей (`max-height: 300px`)
- `.modal-content__actions` — строка кнопок (flex, `justify-content: flex-end`)
- `.modal-content__no-results` — плейсхолдер «ничего не найдено»

### Кнопки

| Класс | Фон | Цвет текста | Применение |
|---|---|---|---|
| `.btn--primary` | `var(--accent)` | `var(--text-on-accent)` | Основное действие |
| `.btn--secondary` | `var(--bg-hover)` | `var(--text-primary)` | Второстепенное (отмена) |
| `.btn--danger` | `var(--danger)` | `var(--text-on-accent)` | Опасное действие (выход, удаление) |

Все кнопки в `.modal-content__actions` имеют отступ `var(--space-md) var(--space-xl)`, `border-radius: var(--radius-md)` и `transition: all var(--transition-base)`. При состоянии `:disabled` — `opacity: 0.6`, курсор `not-allowed`.

### Элементы списка пользователей

`.user-item` — строка пользователя с аватаром и информацией:
- Flex-строка с выравниванием по центру
- Отступ: `var(--space-md) var(--space-lg)`
- Разделитель снизу (кроме последнего): `border-bottom: 1px solid var(--border)`
- При наведении: `background: var(--bg-hover)`

`.user-item__avatar` — круглый аватар `40×40px` с фоном `var(--accent)` и первой буквой имени.

`.user-item__name` — `font-weight: 600`, `color: var(--text-primary)`

`.user-item__email` — `font-size: var(--font-xs)`, `color: var(--text-muted)`

### Стили профиля

`.profile-modal` — модалка профиля фиксированного размера: `420×570px`.

`.profile-modal__header` — шапка с аватаром `56×56px` и именем/email.

`.profile-modal__details` — блок с парами «метка — значение» на фоне `var(--bg-panel)`.

`.profile-modal__value--success` — зелёный цвет `#38a169` (email подтверждён).

`.profile-modal__value--warning` — красный `var(--danger)` (email не подтверждён).

`.profile-modal__section` — сворачиваемая секция на фоне `var(--bg-panel)`.

`.profile-modal__chevron` / `.profile-modal__chevron--open` — анимированная стрелка-индикатор (поворот на 90° при раскрытии).

### Переключатель темы

`.theme-switch__track` — трек `44×24px`, скруглённый (`border-radius: 12px`). В неактивном состоянии фон `var(--border)`, в активном (`.theme-switch__track--on`) — `var(--accent)`.

`.theme-switch__thumb` — ползунок `20×20px` (белый круг). При `.theme-switch__thumb--on` смещается на `translateX(20px)`. Переход — `var(--transition-slow)`.

### Анимации

```css
@keyframes fadeIn {
    from { opacity: 0; }
    to   { opacity: 1; }
}
```

Используется для плавного появления `.modal-overlay`.
