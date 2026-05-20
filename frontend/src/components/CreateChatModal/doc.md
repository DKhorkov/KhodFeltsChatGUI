# Компонент CreateChatModal

## Назначение

Модальное окно создания нового чата. Позволяет выбрать тип (приватный/групповой), ввести название и описание (для группового), найти и выбрать участников.

## Emits

| Событие | Когда |
|---------|-------|
| `close` | Клик на оверлей или кнопку закрытия |
| `chat-created` | Успешное создание чата |

## Inject

- `showError`, `showInfo`

## Ключевое состояние (refs)

| Ref | Описание |
|-----|----------|
| `chatType` | `CHAT_TYPE.PRIVATE` или `CHAT_TYPE.GROUP` |
| `chatTitle` | Название группового чата |
| `chatDescription` | Описание группового чата |
| `searchQuery` | Строка поиска пользователей |
| `searchResults` | Результаты поиска |
| `selectedUserIds` | ID выбранных пользователей |

## Ключевые функции

| Функция | Описание |
|---------|----------|
| `handleSearchInput` | Дебаунс-обёртка над поиском (`SEARCH_DEBOUNCE_MS`) |
| `createChat()` | Валидирует, вызывает `CreateChat(dto)`, эмитит `chat-created` |

## Wails-биндинги

- `create_chat/Handler`: `CreateChat`
- `search_users/Handler`: `SearchUsers`
