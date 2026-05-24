# frontend/src/constants

## Назначение

Все константы фронтенд-приложения. Экспортируются из `index.js`.

## Перечисления (Object.freeze)

| Константа | Значения | Описание |
|-----------|----------|----------|
| `THEME` | `LIGHT: 0`, `DARK: 1` | Типы тем |
| `CHAT_TYPE` | `PRIVATE: 'private'`, `GROUP: 'group'` | Типы чатов |
| `VIEW` | `LOADING`, `LOGIN`, `CHAT` | Экраны приложения |
| `TAB` | `LOGIN`, `REGISTER` | Вкладки экрана входа |
| `WAILS_EVENT` | `NEW_MESSAGE`, `MESSAGE_DELETED`, `CHATS_UPDATED`, `OPEN_CHAT` | События Wails Runtime |

## Скалярные константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `MESSAGES_PAGE_SIZE` | `10` | Размер страницы сообщений |
| `NOTIFICATION_DURATION_MS` | `3000` | Длительность отображения уведомления (мс) |
| `SEARCH_DEBOUNCE_MS` | `500` | Задержка дебаунса поиска (мс) |
| `EMOJI_CLOSE_DELAY_MS` | `500` | Задержка закрытия emoji picker после ухода курсора (мс) |
| `HIGHLIGHT_DURATION_MS` | `1500` | Длительность подсветки сообщения при переходе по ответу (мс) |
