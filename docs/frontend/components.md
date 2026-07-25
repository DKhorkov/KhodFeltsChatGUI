# Компоненты фронтенда

Каждый компонент расположен в `frontend/src/components/<Name>/` и состоит из трёх файлов: `.vue` (шаблон), `.js` (логика), `.css` (стили).

---

## LoginView

Файлы: `LoginView/LoginView.vue`, `LoginView/LoginView.js`

Экран входа и регистрации. Отображается когда `currentView === VIEW.LOGIN`.

### Назначение
Предоставляет две вкладки: «Вход» и «Регистрация». Дополнительно содержит ссылки для повторной отправки письма подтверждения почты и для сброса пароля.

### Props
Нет.

### Emits
| Событие | Когда | Данные |
|---|---|---|
| `login-success` | Успешный вход через `Login()` | — |
| `show-forget-password` | Успешная отправка письма сброса пароля | строка с сообщением для `ForgetPasswordModal` |

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `activeTab` | `ref<string>` | Активная вкладка: `TAB.LOGIN` или `TAB.REGISTER` |
| `loginForm` | `ref<{login, password}>` | Данные формы входа |
| `registerForm` | `ref<{email, username, password, confirmPassword}>` | Данные формы регистрации |

### Inject
- `showError` — отображение ошибки
- `showInfo` — отображение информационного сообщения

### Ключевые функции
- `handleLogin()` — валидирует поля, вызывает `Login(loginForm)`, генерирует `login-success`
- `handleRegister()` — валидирует поля (включая совпадение паролей), вызывает `Register(registerForm)`, после успеха переключает на вкладку входа и подставляет введённые данные
- `sendVerifyEmail()` — вызывает `SendVerifyEmail(loginForm.login)`, показывает информацию через `showInfo`
- `sendForgetPassword()` — вызывает `SendForgetPassword(loginForm.login)`, генерирует `show-forget-password`

### Wails-биндинги
`auth/Handler`: `Login`, `Register`, `SendVerifyEmail`, `SendForgetPassword`

---

## ChatView

Файлы: `ChatView/ChatView.vue`, `ChatView/ChatView.js`

Главный экран приложения с тремя зонами: тулбар, боковая панель со списком чатов, область переписки.

### Назначение
Загружает и отображает чаты текущего пользователя, управляет выбором активного чата, пагинацией сообщений (прокрутка вверх для загрузки старых), отправкой сообщений, real-time обновлениями через Wails Events.

### Props
Нет.

### Emits
| Событие | Когда |
|---|---|
| `logout` | Нажатие кнопки выхода (через ProfileModal) |
| `show-create-chat` | Нажатие кнопки «+» в сайдбаре |
| `show-search-users` | Нажатие кнопки поиска в сайдбаре |
| `show-profile` | Нажатие кнопки профиля в тулбаре |
| `new-message-notification` | Входящее сообщение в неактивный чат; данные: `{text, chatId}` |

### Inject
- `showError`

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `chats` | `ref<Chat[]>` | Список чатов пользователя |
| `selectedChat` | `ref<Chat\|null>` | Активный (открытый) чат |
| `messages` | `ref<Message[]>` | Сообщения активного чата (в хронологическом порядке) |
| `currentUser` | `ref<User\|null>` | Текущий авторизованный пользователь |
| `newMessage` | `ref<string>` | Текст в поле ввода |
| `replyToMessage` | `ref<Message\|null>` | Сообщение, на которое отвечаем |
| `contextMenu` | `ref<Object>` | Состояние контекстного меню (visible, x, y, message, deleteExpanded) |
| `highlightedMessageId` | `ref<number\|null>` | ID подсвеченного сообщения (при скролле к ответу) |
| `messagesListRef` | `ref<HTMLElement>` | Ссылка на контейнер сообщений для скролла |
| `textareaRef` | `ref<HTMLElement>` | Ссылка на textarea для позиционирования курсора после вставки эмодзи |
| `isDarkTheme` | `ref<boolean>` | Состояние темы |
| `isEmojiPickerVisible` | `ref<boolean>` | Видимость панели эмодзи |
| `selectedMember` | `ref<User\|null>` | Участник приватного чата для просмотра профиля |
| `selectedGroupChat` | `ref<Chat\|null>` | Групповой чат для отображения `GroupChatModal` |
| `returnToGroupChat` | `ref<Chat\|null>` | Сохранённый групповой чат для возврата после просмотра профиля участника |

### Локальные (не-реактивные) переменные
- `isLoadingMore: boolean` — флаг загрузки старых сообщений (предотвращает двойной запрос)
- `hasMoreMessages: boolean` — флаг наличия ещё не загруженных сообщений

### Ключевые функции
| Функция | Описание |
|---|---|
| `loadChats()` | Загружает список чатов через `GetUserChats(null)` |
| `loadMessages(chatId)` | Загружает первую страницу сообщений (`MESSAGES_PAGE_SIZE`), разворачивает по времени, скроллит вниз |
| `loadMoreMessages()` | Загружает следующую страницу сообщений по смещению `messages.length`; вызывается при прокрутке к верхней границе |
| `selectChat(chat)` | Устанавливает `selectedChat`, загружает сообщения, помечает чат как прочитанный |
| `openChatById(chatId)` | Находит чат по ID в `chats` и вызывает `selectChat`. Используется из `App.js` при клике по уведомлению |
| `sendMessage()` | Отправляет сообщение через `SendMessage(chatId, text, replyId)`, помечает все прочитанными. UI обновляется по WS-событию от сервера |
| `handleNewMessage(message)` | Обработчик события `new_message`: добавляет в открытый чат или показывает уведомление |
| `handleMessageDeleted(payload)` | Обработчик события `message_deleted`: удаляет сообщение из UI по ID |
| `handleChatsUpdated(updatedChats)` | Обработчик события `chats_updated`: заменяет весь список чатов |
| `deleteContextMessage(forAll)` | Удаляет сообщение через `DeleteMessage()`. UI обновляется по WS-событию от сервера |
| `replyToContextMessage()` | Устанавливает `replyToMessage` для ответа через контекстное меню |
| `openContextMenu(event, message)` | Открывает контекстное меню сообщения с позиционированием к viewport |
| `copyContextMessage()` | Копирует текст сообщения в буфер обмена |
| `getChatTitle(chat)` | Возвращает `chat.title` или имя собеседника (для приватного), или `Чат #id` |
| `getOtherMember(chat)` | Для приватного чата возвращает участника, который не является текущим пользователем |
| `openMemberProfile(chat)` | Открывает профиль участника (приватный чат) или `GroupChatModal` (групповой) |
| `openGroupMemberProfile(member)` | Открывает профиль участника группового чата; сохраняет группу в `returnToGroupChat` |
| `closeMemberProfile()` | Закрывает профиль; если открыт из группового чата — возвращает в `GroupChatModal` |
| `insertEmoji(emoji)` | Вставляет эмодзи в позицию курсора в textarea |
| `toggleTheme()` | Вызывает `ToggleTheme()`, обновляет `isDarkTheme` и атрибут `data-bs-theme` на `<html>` |
| `isFirstUnread(message, index)` | Определяет, является ли сообщение первым непрочитанным (для разделителя «Новые сообщения») |
| `formatTime(dateStr)` | Форматирует дату/время в локаль `ru-RU` |
| `formatDate(dateStr)` | Форматирует дату в локаль `ru-RU` (день, месяц, год) |

### Жизненный цикл
- `onMounted`: загружает текущего пользователя, тему, настройки, чаты, вызывает `StartListening()` для подключения к WebSocket на Go-стороне, подписывается на Wails Events (`NEW_MESSAGE`, `MESSAGE_DELETED`, `CHATS_UPDATED`, `OPEN_CHAT`)
- `onUnmounted`: вызывает `StopListening()`, отписывается от Events
- `watch(messagesListRef)`: при появлении DOM-элемента навешивает scroll-обработчик для бесконечной пагинации вверх

### Дочерние компоненты
- `EmojiPicker` — отображается в области ввода сообщения
- `GroupChatModal` — отображается поверх при просмотре информации о групповом чате

### Wails-биндинги
`chats/Handler`: `GetUserChats`, `StartListening`, `StopListening`
`messages/Handler`: `GetChatMessages`, `SendMessage`, `DeleteMessage`
`users/Handler`: `GetCurrentUser`
`theme/Handler`: `GetTheme`, `ToggleTheme`

---

## AlertModal

Файлы: `AlertModal/AlertModal.vue`, `AlertModal/AlertModal.js`

Модальное окно для отображения ошибок и информационных сообщений.

### Назначение
Блокирующий диалог с одной кнопкой «OK». Закрывается кликом на оверлей, нажатием Escape или кнопкой OK.

### Props
| Prop | Тип | Обязательный | Описание |
|---|---|---|---|
| `message` | `String` | да | Текст сообщения |
| `type` | `String` | нет (по умолчанию `'info'`) | Тип: `'error'` или `'info'` |

### Emits
| Событие | Когда |
|---|---|
| `close` | Клик на оверлей, Escape, нажатие OK |

### Особенности
Использует директиву `v-focus` для захвата фокуса. В зависимости от `type` показывает иконку, заголовок («Ошибка» / «Информация») и CSS-класс.

---

## NotificationToast

Файлы: `NotificationToast/NotificationToast.vue`, `NotificationToast/NotificationToast.js`

Всплывающее уведомление в стопке уведомлений.

### Назначение
Эфемерное уведомление (показывается `NOTIFICATION_DURATION_MS` мс). Клик по уведомлению переключает на связанный чат. Стопка уведомлений управляется в `App.js`.

### Props
| Prop | Тип | Обязательный | Описание |
|---|---|---|---|
| `message` | `String` | да | Текст уведомления |

### Emits
| Событие | Когда |
|---|---|
| `click` | Клик на тело уведомления |
| `close` | Клик на кнопку закрытия (`×`) |

---

## CreateChatModal

Файлы: `CreateChatModal/CreateChatModal.vue`, `CreateChatModal/CreateChatModal.js`

Модальное окно создания нового чата.

### Назначение
Позволяет выбрать тип чата (приватный или групповой), ввести название/описание (только для группового), найти участников по имени пользователя через поиск с дебаунсом, выбрать несколько участников (checkbox) и создать чат.

### Props
Нет.

### Emits
| Событие | Когда |
|---|---|
| `close` | Клик на оверлей или кнопку закрытия |
| `chat-created` | Успешное создание чата через `CreateChat()` |

### Inject
- `showError`, `showInfo`

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `chatType` | `ref<string>` | `CHAT_TYPE.PRIVATE` или `CHAT_TYPE.GROUP` |
| `chatTitle` | `ref<string>` | Название группового чата |
| `chatDescription` | `ref<string>` | Описание группового чата |
| `searchQuery` | `ref<string>` | Строка поиска пользователей |
| `searchResults` | `ref<User[]>` | Результаты поиска |
| `selectedUserIds` | `ref<number[]>` | ID выбранных пользователей |

### Ключевые функции
- `searchUsers()` — вызывает `SearchUsers({username: searchQuery}, null)`
- `handleSearchInput` — дебаунс-обёртка над `searchUsers` с задержкой `SEARCH_DEBOUNCE_MS`
- `createChat()` — валидирует выбор участников, вызывает `CreateChat({type, memberIDs, title, description})`, генерирует `chat-created`

### Wails-биндинги
`chats/Handler`: `CreateChat`
`users/Handler`: `SearchUsers`

---

## GroupChatModal

Файлы: `GroupChatModal/GroupChatModal.vue`, `GroupChatModal/GroupChatModal.js`

Модальное окно с информацией о групповом чате и списком его участников.

### Назначение
Отображает аватар (первая буква названия), название, описание и список участников группового чата. Каждый участник кликабелен для просмотра его профиля.

### Props
| Prop | Тип | Обязательный | Описание |
|---|---|---|---|
| `chat` | `Object` | да | Объект чата типа `Chat` |
| `currentUser` | `Object` | да | Текущий пользователь (для пометки «вы») |

### Emits
| Событие | Когда | Данные |
|---|---|---|
| `close` | Клик на оверлей или кнопку закрытия | — |
| `open-member-profile` | Клик на участника | объект `User` |

### Вычисляемые свойства (computed)
- `chatTitle` — `chat.title` или `Чат #id`
- `chatInitial` — первая буква `chatTitle` в верхнем регистре
- `membersCount` — `chat.members.length`

### Ключевые функции
- `openMemberProfile(member)` — генерирует событие `open-member-profile`

---

## SearchUsersModal

Файлы: `SearchUsersModal/SearchUsersModal.vue`, `SearchUsersModal/SearchUsersModal.js`

Модальное окно поиска пользователей по имени.

### Назначение
Позволяет найти пользователей по `username` с дебаунсом. При клике на пользователя отображается его профиль прямо внутри модалки (без открытия нового окна). Кнопка «×» в режиме профиля возвращает к списку результатов.

### Props
Нет.

### Emits
| Событие | Когда |
|---|---|
| `close` | Клик на оверлей или кнопку закрытия |

### Inject
- `showError`

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `searchQuery` | `ref<string>` | Строка поиска |
| `searchResults` | `ref<User[]>` | Результаты поиска |
| `hasSearched` | `ref<boolean>` | Был ли выполнен хотя бы один поиск (для отображения «не найдено») |
| `selectedUser` | `ref<User\|null>` | Выбранный пользователь для просмотра профиля |

### Ключевые функции
- `searchUsers()` — вызывает `SearchUsers({username: searchQuery}, null)`, устанавливает `hasSearched = true`
- `handleSearchInput` — дебаунс-обёртка с задержкой `SEARCH_DEBOUNCE_MS`
- `formatDate(dateStr)` — форматирует дату регистрации в локаль `ru-RU`

### Wails-биндинги
`users/Handler`: `SearchUsers`

---

## ForgetPasswordModal

Файлы: `ForgetPasswordModal/ForgetPasswordModal.vue`, `ForgetPasswordModal/ForgetPasswordModal.js`

Модальное окно сброса пароля.

### Назначение
Отображается после отправки письма со ссылкой сброса пароля. Пользователь вводит код из письма, новый пароль и его подтверждение.

### Props
| Prop | Тип | Обязательный | Описание |
|---|---|---|---|
| `message` | `String` | да | Информационное сообщение (адрес, на который отправлено письмо) |

### Emits
| Событие | Когда |
|---|---|
| `close` | Успешный сброс пароля или нажатие «Назад» |

### Inject
- `showError`, `showInfo`

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `token` | `ref<string>` | Код из письма |
| `newPassword` | `ref<string>` | Новый пароль |
| `confirmPassword` | `ref<string>` | Подтверждение пароля |

### Ключевые функции
- `resetPassword()` — валидирует поля и совпадение паролей, вызывает `ForgetPassword(token, {newPassword})`, показывает подтверждение и закрывает модалку

### Wails-биндинги
`auth/Handler`: `ForgetPassword`

---

## ProfileModal

Файлы: `ProfileModal/ProfileModal.vue`, `ProfileModal/ProfileModal.js`

Модальное окно профиля текущего пользователя.

### Назначение
Отображает информацию о текущем пользователе (имя, email, статус подтверждения email, дата регистрации). Содержит три сворачиваемые секции: редактирование профиля (смена username), смена пароля и переключатель темы. Также кнопка выхода из аккаунта.

### Props
| Prop | Тип | Обязательный | Описание |
|---|---|---|---|
| `user` | `Object` | да | Текущий пользователь (`User`) |
| `isDarkTheme` | `Boolean` | да | Состояние тёмной темы для переключателя |

### Emits
| Событие | Когда |
|---|---|
| `close` | Клик на оверлей или кнопку закрытия |
| `toggle-theme` | Клик на переключатель темы |
| `logout` | Клик на «Выйти из аккаунта» |
| `user-updated` | Успешное обновление профиля; данные: обновлённый объект `User` |

### Inject
- `showError`, `showInfo`

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `isEditProfileOpen` | `ref<boolean>` | Открыта ли секция редактирования профиля |
| `editUsername` | `ref<string>` | Новое имя пользователя |
| `isChangePasswordOpen` | `ref<boolean>` | Открыта ли секция смены пароля |
| `oldPassword` | `ref<string>` | Текущий пароль |
| `newPassword` | `ref<string>` | Новый пароль |
| `confirmPassword` | `ref<string>` | Подтверждение нового пароля |

### Ключевые функции
- `changePassword()` — валидирует поля и совпадение паролей, вызывает `ChangePassword({oldPassword, newPassword})`, сбрасывает форму
- `updateUser()` — вызывает `UpdateUser({username: editUsername})`, генерирует `user-updated` с возвращённым объектом пользователя
- `formatDate(dateStr)` — форматирует дату в локаль `ru-RU`

### Wails-биндинги
`auth/Handler`: `ChangePassword`
`users/Handler`: `UpdateUser`

---

## PasswordInput

Файлы: `PasswordInput/PasswordInput.vue`, `PasswordInput/PasswordInput.js`, `PasswordInput/PasswordInput.css`

Поле ввода пароля с кнопкой-глазом. Клик по кнопке переключает `type` инпута между `password` и `text`.

### Назначение
Используется в `LoginView` (логин + регистрация), `ForgetPasswordModal` (сброс пароля) и `ProfileModal` (смена пароля).

### Props
| Проп | Тип | По умолчанию | Описание |
|---|---|---|---|
| `modelValue` | `String` | `''` | Значение поля (для `v-model`) |
| `placeholder` | `String` | `''` | Плейсхолдер |
| `id` | `String` | `''` | HTML id инпута |
| `name` | `String` | `''` | HTML name инпута |
| `required` | `Boolean` | `false` | Флаг обязательности |
| `autocomplete` | `String` | `'current-password'` | HTML autocomplete (обычно `current-password` или `new-password`) |
| `inputClass` | `String` | `''` | Дополнительный класс для инпута (для интеграции в существующие формы, например `login-form__input`) |

### Emits
| Событие | Когда | Данные |
|---|---|---|
| `update:modelValue` | Ввод в поле | новое значение |

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `visible` | `ref<Boolean>` | Виден ли пароль в открытом виде |

### Ключевые функции
- `toggle()` — инвертирует `visible`
- `onInput(event)` — эмитит `update:modelValue` со значением из `event.target.value`

## EmojiPicker

Файлы: `EmojiPicker/EmojiPicker.vue`, `EmojiPicker/EmojiPicker.js`

Панель выбора эмодзи.

### Назначение
Отображается поверх поля ввода сообщения в `ChatView`. Содержит 4 категории эмодзи с переключением по вкладкам.

### Props
Нет.

### Emits
| Событие | Когда | Данные |
|---|---|---|
| `select` | Клик на эмодзи | строка эмодзи |

### Категории (определены в модуле, вне компонента)
| Ключ | Метка | Количество |
|---|---|---|
| `smileys` | Смайлы | 56 |
| `gestures` | Жесты | 32 |
| `hearts` | Сердца | 24 |
| `objects` | Объекты | 32 |

### Ключевое состояние (refs)
| Переменная | Тип | Описание |
|---|---|---|
| `activeCategory` | `ref<string>` | Имя активной категории (по умолчанию `'smileys'`) |

### Вычисляемые свойства (computed)
- `activeEmojis` — массив эмодзи активной категории
