# Архитектура фронтенда

## Стек технологий

- **Vue 3** (Composition API, `setup()`) — основной UI-фреймворк
- **Vite** — инструмент сборки и локального dev-сервера
- Без роутера (нет Vue Router): переключение представлений реализовано вручную через реактивную переменную `currentView`
- Без глобального стора (нет Pinia/Vuex): состояние передаётся через props, events и provide/inject

## Управление состоянием

Проект использует три механизма Vue без внешних библиотек:

1. **Props / emit** — основной способ передачи данных вниз и событий наверх между родителем и дочерними компонентами.
2. **provide / inject** — в `App.js` предоставляются (`provide`) два глобальных хелпера:
   - `showError(message)` — показывает `AlertModal` с типом `error`
   - `showInfo(message)` — показывает `AlertModal` с типом `info`
   
   Любой дочерний компонент получает их через `inject('showError')` / `inject('showInfo')` без необходимости передавать пропсы через несколько уровней.
3. **Template ref** — `App.vue` держит `chatViewRef` (ссылку на экземпляр `ChatView`) и вызывает его публичные методы напрямую (`chatViewRef.value.loadChats()`, `chatViewRef.value.openChatById(id)`).

Подробное описание паттернов — в `docs/frontend/provide_inject_emit.md`.

## Структура компонентов

Каждый компонент оформлен как отдельная папка с тремя файлами:

```
src/components/ComponentName/
    ComponentName.vue   — шаблон (использует <script src="./ComponentName.js">)
    ComponentName.js    — логика (Composition API setup())
    ComponentName.css   — локальные стили
```

Такое разделение позволяет работать с шаблоном, логикой и стилями в отдельных файлах без использования однофайловых SFC с `<script setup>`.

## Точка входа

Файл: `frontend/src/main.js`

```js
const app = createApp(App)

// Глобальная директива v-focus: при монтировании элемента вызывает el.focus()
// Используется в модальных окнах для автоматического перехвата фокуса и обработки Escape
app.directive('focus', {
    mounted: (el) => el.focus(),
})

app.mount('#app')
```

Импортируется `global.css` (дизайн-токены и общие стили).

## App.vue / App.js — корневой компонент

Файлы: `frontend/src/App.vue`, `frontend/src/App.js`

Корневой компонент реализует:

- **Переключение представлений** через `currentView` (тип `ref<string>`), значения из константы `VIEW`:
  - `VIEW.LOADING` — начальное состояние при проверке сессии
  - `VIEW.LOGIN` — экран авторизации (`LoginView`)
  - `VIEW.CHAT` — главный экран чата (`ChatView`)
- **Проверку сессии** при монтировании: вызывает `Authenticate()` из `auth/Handler`. При успехе переключает на `VIEW.CHAT`, при ошибке — на `VIEW.LOGIN`.
- **Загрузку темы** при монтировании через `GetTheme()` из `theme/Handler`. Тема применяется через `document.documentElement.setAttribute('data-bs-theme', ...)`.
- **Глобальные модалки**: `CreateChatModal`, `SearchUsersModal`, `ForgetPasswordModal`, `ProfileModal`, `AlertModal` — управляются булевыми ref-переменными (`isCreateChatVisible`, `isSearchUsersVisible`, и т. д.).
- **Стек уведомлений**: массив `notifications` (ref), каждое уведомление содержит `{id, message, chatId}`. Уведомления автоматически удаляются через `NOTIFICATION_DURATION_MS` мс. Клик по уведомлению с `chatId` переключает активный чат через `chatViewRef.value.openChatById(chatId)`.
- **Provide** для `showError` / `showInfo`.

## Взаимодействие с Go-бэкендом через Wails (IPC)

Wails предоставляет механизм межпроцессного взаимодействия (IPC) между фронтендом (JavaScript) и бэкендом (Go).

Авто-сгенерированные биндинги расположены в `frontend/wailsjs/go/<пакет>/Handler.js`. Каждая экспортируемая функция Go-хендлера становится асинхронной JS-функцией через глобальный объект `window['go']`:

```js
// Пример из wailsjs/go/auth/Handler.js
export function Login(arg1) {
    return window['go']['auth']['Handler']['Login'](arg1);
}
```

Фронтенд импортирует функции напрямую:

```js
import { Login, Logout } from '../../../wailsjs/go/auth/Handler'
```

Все вызовы возвращают `Promise`. При ошибке промис отклоняется со строкой ошибки, которая передаётся в `showError(err)`.

## Реальное время: EventsOn / EventsEmit

Wails Runtime экспонирует `window.runtime` с событийной шиной. Фронтенд использует её для получения push-уведомлений от Go-бэкенда.

Подписка (в `ChatView.js`, `onMounted`):

```js
window.runtime.EventsOn(WAILS_EVENT.NEW_MESSAGE, handleNewMessage)
window.runtime.EventsOn(WAILS_EVENT.CHATS_UPDATED, handleChatsUpdated)
```

Отписка (в `ChatView.js`, `onUnmounted`):

```js
window.runtime.EventsOff(WAILS_EVENT.NEW_MESSAGE)
window.runtime.EventsOff(WAILS_EVENT.CHATS_UPDATED)
```

Поддерживаемые события (константы из `src/constants/index.js`):

| Константа | Строковое значение | Payload | Обработчик |
|---|---|---|---|
| `WAILS_EVENT.NEW_MESSAGE` | `'new_message'` | объект `Message` | `handleNewMessage` |
| `WAILS_EVENT.CHATS_UPDATED` | `'chats_updated'` | массив `Chat[]` | `handleChatsUpdated` |

При получении `new_message`: если сообщение относится к открытому чату — добавляется в список сообщений. Иначе — обновляется список чатов и показывается уведомление.

При получении `chats_updated`: список чатов заменяется целиком.

## CSS: глобальные стили и дизайн-токены

Файл: `frontend/src/styles/global.css`

Все визуальные константы вынесены в CSS custom properties (переменные) в корневом селекторе `:root`. Тёмная тема переопределяет их в `[data-bs-theme="dark"]`. Переключение темы: бэкенд сохраняет выбор, фронтенд устанавливает атрибут на `<html>`.

Подробнее — в `docs/frontend/styles.md`.
