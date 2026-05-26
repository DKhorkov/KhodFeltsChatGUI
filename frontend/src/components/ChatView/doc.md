# Компонент ChatView

## Назначение

Главный экран приложения с тремя зонами: тулбар, боковая панель со списком чатов, область переписки.

## Emits

| Событие | Когда |
|---------|-------|
| `logout` | Выход через ProfileModal |
| `show-create-chat` | Клик «+» в сайдбаре |
| `show-search-users` | Клик на поиск в сайдбаре |
| `show-profile` | Клик на профиль в тулбаре |
| `new-message-notification` | Входящее сообщение в неактивный чат |

## Inject

- `showError`

## Ключевое состояние (refs)

| Ref | Тип | Описание |
|-----|-----|----------|
| `chats` | `Chat[]` | Список чатов пользователя |
| `selectedChat` | `Chat\|null` | Активный чат |
| `messages` | `Message[]` | Сообщения активного чата |
| `currentUser` | `User\|null` | Текущий пользователь |
| `newMessage` | `string` | Текст в поле ввода |
| `isDarkTheme` | `boolean` | Состояние темы |
| `isEmojiPickerVisible` | `boolean` | Видимость панели эмодзи |
| `selectedMember` | `User\|null` | Участник для просмотра профиля |
| `selectedGroupChat` | `Chat\|null` | Групповой чат для GroupChatModal |
| `webPushConsents` | `number` | Битовая маска согласий на уведомления |
| `editingMessage` | `Message\|null` | Сообщение, которое редактируется в данный момент |

## Ключевые функции

| Функция | Описание |
|---------|----------|
| `loadChats()` | Загружает список чатов через `GetUserChats(null)` |
| `loadMessages(chatId)` | Загружает первую страницу сообщений, скроллит вниз |
| `loadMoreMessages()` | Пагинация вверх при прокрутке к верхней границе |
| `selectChat(chat)` | Устанавливает активный чат, загружает сообщения |
| `sendMessage()` | Отправляет сообщение через `SendMessage()` или редактирует через `UpdateMessage()` (если `editingMessage` установлен), помечает все прочитанными. UI обновляется по WS-событию от сервера |
| `handleNewMessage(msg)` | Обработчик Wails-события `new_message`: добавляет в открытый чат или показывает уведомление |
| `handleMessageDeleted(payload)` | Обработчик Wails-события `message_deleted`: удаляет сообщение из UI. Логирует предупреждение если ID не найден |
| `handleMessageEdited(payload)` | Обработчик Wails-события `message_edited`: получает обновлённое сообщение через `GetMessageByID()` и обновляет текст и `updatedAt` в UI |
| `handleChatsUpdated(chats)` | Обработчик Wails-события `chats_updated` |
| `deleteContextMessage(forAll)` | Удаляет сообщение через `DeleteMessage()`. UI обновляется по WS-событию `message_deleted` от сервера |
| `editContextMessage()` | Устанавливает `editingMessage` для редактирования сообщения из контекстного меню (только для своих сообщений) |
| `cancelEdit()` | Сбрасывает `editingMessage` и очищает поле ввода, отменяя режим редактирования |
| `replyToContextMessage()` | Устанавливает `replyToMessage` для ответа на сообщение из контекстного меню |
| `openContextMenu(event, message)` | Открывает контекстное меню с позиционированием (зажимает к viewport) |
| `getLastMessagePreview(chat)` | Формирует превью последнего сообщения для списка чатов |
| `getChatTitle(chat)` | Возвращает название чата или имя собеседника |
| `toggleTheme()` | Переключает тему через `ToggleTheme()` |
| `insertEmoji(emoji)` | Вставляет эмодзи в позицию курсора |

## Дочерние компоненты

- `EmojiPicker` — панель выбора эмодзи
- `GroupChatModal` — информация о групповом чате

## Wails-биндинги

- `chats/Handler`: `GetUserChats`, `StartListening`, `StopListening`
- `messages/Handler`: `GetChatMessages`, `GetMessageByID`, `SendMessage`, `DeleteMessage`, `UpdateMessage`
- `users/Handler`: `GetCurrentUser`
- `theme/Handler`: `GetTheme`, `ToggleTheme`
- `settings/Handler`: `GetSettings`
- `notification/Handler`: `ShowNotification`

## Wails Events

- `NEW_MESSAGE` — новое сообщение через WebSocket (единственный источник UI-обновлений при отправке)
- `MESSAGE_DELETED` — удаление сообщения через WebSocket (единственный источник UI-обновлений при удалении)
- `MESSAGE_EDITED` — редактирование сообщения через WebSocket (единственный источник UI-обновлений при редактировании)
- `CHATS_UPDATED` — обновление списка чатов (каждые 5 сек с бэкенда)
- `OPEN_CHAT` — открытие чата по клику на системное уведомление

## UI: редактирование сообщений

- Контекстное меню содержит пункт "Редактировать" (только для своих сообщений)
- При редактировании в composer появляется edit bar (аналогичен reply bar, с orange/warning акцентом)
- Кнопка отправки показывает "Сохранить" в режиме редактирования
- CSS-классы: `.conversation__edit-bar`, `.conversation__edit-bar-content`, `.conversation__edit-bar-label`, `.conversation__edit-bar-text`, `.conversation__edit-bar-close`

## Архитектурный принцип

Клиент не обновляет UI самостоятельно при отправке/удалении/редактировании сообщений. Команды отправляются на сервер, сервер рассылает WS-события всем участникам чата (включая отправителя), слушатели событий обновляют UI (single source of truth).
