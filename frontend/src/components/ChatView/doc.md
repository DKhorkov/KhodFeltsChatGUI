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
| `avatarZoomSrc` | `string\|null` | URL аватара для увеличенного просмотра |
| `selectedGroupChat` | `Chat\|null` | Групповой чат для GroupChatModal |
| `webPushConsents` | `number` | Битовая маска согласий на уведомления |
| `editingMessage` | `Message\|null` | Сообщение, которое редактируется в данный момент |
| `unreadMessageIds` | `Set<number>` | Реактивный Set ID сообщений, прилетевших с момента отскролла вверх. Хранится как Set (а не число), чтобы при `MESSAGE_DELETED` корректно вычесть удалённое сообщение из бейджа. Шаблон биндится на `.size` напрямую |
| `isAtBottom` | `boolean` | Видно ли последнее сообщение чата (`IntersectionObserver` на последнем `.message-bubble`, threshold 0.1) |

## Ключевые функции

| Функция | Описание |
|---------|----------|
| `loadChats()` | Загружает список чатов через `GetUserChats(null)` и пересчитывает счётчик непрочитанных через `updateUnreadBadge()` |
| `updateUnreadBadge(chats)` | Считает `Σ chat.unreadCount` и публикует его в title окна через биндинг `SetUnreadBadge(total)`. Единственная точка обновления счётчика на клиенте |
| `loadMessages(chatId)` | Загружает первую страницу сообщений, скроллит вниз |
| `loadMoreMessages()` | Пагинация вверх при прокрутке к верхней границе |
| `selectChat(chat)` | Устанавливает активный чат, загружает сообщения |
| `sendMessage()` | Отправляет сообщение через `SendMessage()` или редактирует через `UpdateMessage()` (если `editingMessage` установлен), помечает все прочитанными. UI обновляется по WS-событию от сервера |
| `handleNewMessage(msg)` | Обработчик Wails-события `new_message`. В открытом чате: снимок `wasAtBottom` ДО `messages.push`. Если своё сообщение или `wasAtBottom` — `scrollToBottom()` (мгновенно, follow mode); иначе — `unreadMessageIds.add(msg.id)` без скролла. В неактивном чате — emit `new-message-notification`. Системное OS-уведомление через `ShowNotification` показывается всегда, кроме случая «окно в фокусе И активный чат = чат с новым сообщением». Гейтится `webPushConsents & CONSENT_NEW_MESSAGE` |
| `handleMessageDeleted(payload)` | Обработчик Wails-события `message_deleted`: удаляет сообщение из UI и вычитает его из `unreadMessageIds` (если оно там было). Логирует предупреждение если ID не найден |
| `handleMessageEdited(payload)` | Обработчик Wails-события `message_edited`: получает обновлённое сообщение через `GetMessageByID()` и обновляет текст и `updatedAt` в UI |
| `handleChatsUpdated(chats)` | Обработчик Wails-события `chats_updated`. Обновляет `chats.value` и пересчитывает счётчик через `updateUnreadBadge(updatedChats)` |
| `deleteContextMessage(forAll)` | Удаляет сообщение через `DeleteMessage()`. UI обновляется по WS-событию `message_deleted` от сервера |
| `editContextMessage()` | Устанавливает `editingMessage` для редактирования сообщения из контекстного меню (только для своих сообщений) |
| `cancelEdit()` | Сбрасывает `editingMessage` и очищает поле ввода, отменяя режим редактирования |
| `replyToContextMessage()` | Устанавливает `replyToMessage` для ответа на сообщение из контекстного меню |
| `openContextMenu(event, message)` | Открывает контекстное меню с позиционированием (зажимает к viewport) |
| `getLastMessagePreview(chat)` | Формирует превью последнего сообщения для списка чатов |
| `getChatTitle(chat)` | Возвращает название чата или имя собеседника |
| `getChatAvatarPath(chat)` | Возвращает URL аватара собеседника (для приватных чатов) или null |
| `openAvatarZoom(src)` | Открывает оверлей увеличенного просмотра аватара |
| `toggleTheme()` | Переключает тему через `ToggleTheme()` |
| `insertEmoji(emoji)` | Вставляет эмодзи в позицию курсора |
| `onScrollDownClick()` | Smooth scroll к низу через `scrollTo({ behavior: 'smooth' })`. Сброс счётчика и скрытие кнопки происходят через IntersectionObserver, когда последнее сообщение становится видимым |
| `resetScrollDownState()` | Очищает `unreadMessageIds`, `isAtBottom = true`, `observer.disconnect()`. Вызывается в `closeChat()` |

## Дочерние компоненты

- `EmojiPicker` — панель выбора эмодзи
- `GroupChatModal` — информация о групповом чате

## Wails-биндинги

- `chats/Handler`: `GetUserChats`, `StartListening`, `StopListening`
- `messages/Handler`: `GetChatMessages`, `GetMessageByID`, `SendMessage`, `DeleteMessage`, `UpdateMessage`
- `users/Handler`: `GetCurrentUser`
- `theme/Handler`: `GetTheme`, `ToggleTheme`
- `settings/Handler`: `GetSettings`
- `notification/Handler`: `ShowNotification`, `SetUnreadBadge`

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

## UI: кнопка «к последнему сообщению»

- Появляется, когда последнее сообщение чата вне видимой области (`IntersectionObserver` на последнем `.message-bubble`, threshold 0.1).
- Полупрозрачная круглая кнопка 44px в правом нижнем углу области сообщений (позиционирована относительно `.conversation__composer` через `bottom: 100%` + `margin-bottom`).
- При входящем чужом сообщении в открытом чате и `isAtBottom === false` — автоскролл не выполняется, инкрементируется `unreadCount`. Бейдж: число до 99, дальше «99+». Своё сообщение и follow mode (`wasAtBottom`) — мгновенный `scrollToBottom()`.
- Клик: smooth scroll к низу через `scrollTo({ behavior: 'smooth' })`. Existing `scrollToBottom()` (мгновенный) сохраняется для открытия чата, отправки своего сообщения и follow-mode.
- Observer создаётся в `watch(messagesListRef, ...)` (с cleanup). Переподписка на новый последний bubble — через `watch(() => messages.value.length, ..., { flush: 'post' })`. Источник — длина массива (а не сам `messages`), чтобы триггер срабатывал на `push`/`splice` (новое или удалённое сообщение), но не на мутации полей внутри элементов (`messages[idx].text = ...` в `handleMessageEdited`).
- Сброс состояния в `closeChat()` через `resetScrollDownState()`.
- CSS-классы: `.conversation__scroll-down`, `.conversation__scroll-down-icon`, `.conversation__scroll-down-badge`.

## Архитектурный принцип

Клиент не обновляет UI самостоятельно при отправке/удалении/редактировании сообщений. Команды отправляются на сервер, сервер рассылает WS-события всем участникам чата (включая отправителя), слушатели событий обновляют UI (single source of truth).
