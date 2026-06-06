# Компонент ConfirmDeleteModal

## Назначение

Модальное окно подтверждения с двумя кнопками: «Отмена» и подтверждающая (по умолчанию «Удалить»). Используется для опасных действий, требующих явного подтверждения пользователя.

## Props

| Prop | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| `message` | `String` | — (обязательный) | Текст подтверждения |
| `title` | `String` | `'Подтверждение'` | Заголовок модалки |
| `confirmText` | `String` | `'Удалить'` | Текст подтверждающей кнопки |
| `cancelText` | `String` | `'Отмена'` | Текст кнопки отмены |
| `confirmType` | `String` | `'danger'` | Стиль подтверждающей кнопки: `danger` или `primary` |

## Emits

| Событие | Когда |
|---------|-------|
| `confirm` | Клик на подтверждающую кнопку |
| `cancel` | Клик на «Отмена», на оверлей, или нажатие Escape |

## Стили

- Использует общие классы `alert-modal`, `modal-overlay--alert` из AlertModal/global.css.
- Дополнительные классы (в `ConfirmDeleteModal.css`):
  - `.alert-modal__actions` — ряд из двух кнопок
  - `.alert-modal__btn--danger` / `.alert-modal__btn--primary` / `.alert-modal__btn--secondary`

## Использование

```vue
<ConfirmDeleteModal
  v-if="isConfirmOpen"
  message="Вы уверены?"
  @confirm="handleConfirm"
  @cancel="isConfirmOpen = false"
/>
```
