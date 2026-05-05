# Константы фронтенда

Файл: `frontend/src/constants/index.js`

Все константы заморожены через `Object.freeze()` — это предотвращает случайное изменение значений в рантайме.

---

## THEME

Тип темы оформления. Используется в `App.js`, `ChatView.js`, `ProfileModal.js`.

```js
export const THEME = Object.freeze({
    LIGHT: 0,
    DARK:  1,
})
```

| Ключ | Значение | Описание |
|---|---|---|
| `LIGHT` | `0` | Светлая тема |
| `DARK` | `1` | Тёмная тема |

Применение: результат `GetTheme()` / `ToggleTheme()` сравнивается с `THEME.DARK` для установки `isDarkTheme`. При тёмной теме на элемент `<html>` ставится `data-bs-theme="dark"`, при светлой — `data-bs-theme="light"`.

---

## CHAT_TYPE

Тип чата. Используется в `ChatView.js`, `CreateChatModal.js`.

```js
export const CHAT_TYPE = Object.freeze({
    PRIVATE: 'private',
    GROUP:   'group',
})
```

| Ключ | Значение | Описание |
|---|---|---|
| `PRIVATE` | `'private'` | Приватный чат (два участника) |
| `GROUP` | `'group'` | Групповой чат |

Применение: при создании чата передаётся в `CreateChat({type: chatType})`. В `ChatView` используется для определения отображения — заголовка чата (`getChatTitle`) и выбора модалки при клике на аватар чата (`openMemberProfile`).

---

## VIEW

Идентификатор текущего представления. Управляет тем, какой корневой экран отображается в `App.vue`.

```js
export const VIEW = Object.freeze({
    LOADING: 'loading',
    LOGIN:   'login',
    CHAT:    'chat',
})
```

| Ключ | Значение | Описание |
|---|---|---|
| `LOADING` | `'loading'` | Начальное состояние (проверка сессии); экран не показывается |
| `LOGIN` | `'login'` | Экран входа / регистрации (`LoginView`) |
| `CHAT` | `'chat'` | Основной экран чата (`ChatView`) |

---

## TAB

Идентификатор вкладки на экране входа. Используется только в `LoginView.js`.

```js
export const TAB = Object.freeze({
    LOGIN:    'login',
    REGISTER: 'register',
})
```

| Ключ | Значение | Описание |
|---|---|---|
| `LOGIN` | `'login'` | Вкладка «Вход» |
| `REGISTER` | `'register'` | Вкладка «Регистрация» |

---

## WAILS_EVENT

Имена событий Wails Runtime для подписки на push-уведомления от Go-бэкенда. Используется в `ChatView.js` и `src/constants/index.js`.

```js
export const WAILS_EVENT = Object.freeze({
    NEW_MESSAGE:    'new_message',
    CHATS_UPDATED:  'chats_updated',
})
```

| Ключ | Значение | Инициатор | Данные |
|---|---|---|---|
| `NEW_MESSAGE` | `'new_message'` | Go-бэкенд (при поступлении нового сообщения) | объект `Message` |
| `CHATS_UPDATED` | `'chats_updated'` | Go-бэкенд (при изменении состава чатов) | массив `Chat[]` |

---

## MESSAGES_PAGE_SIZE

```js
export const MESSAGES_PAGE_SIZE = 10
```

Количество сообщений, запрашиваемых за один запрос. Используется в `ChatView.js` в `loadMessages()` и `loadMoreMessages()`. Если сервер вернул меньше `MESSAGES_PAGE_SIZE` сообщений, флаг `hasMoreMessages` сбрасывается в `false` и дальнейшие запросы не выполняются.

---

## NOTIFICATION_DURATION_MS

```js
export const NOTIFICATION_DURATION_MS = 3000
```

Время отображения всплывающего уведомления (`NotificationToast`) в миллисекундах. Используется в `App.js` в `addNotification()` — через `setTimeout` уведомление удаляется из стека.

---

## SEARCH_DEBOUNCE_MS

```js
export const SEARCH_DEBOUNCE_MS = 500
```

Задержка дебаунса при вводе в поле поиска пользователей в миллисекундах. Используется в `CreateChatModal.js` и `SearchUsersModal.js`. Поиск выполняется только если пользователь не вводил текст в течение 500 мс.
