# Компонент ProfileModal

## Назначение

Модальное окно профиля текущего пользователя: информация, аватар (загрузка/удаление), редактирование имени, смена пароля, переключатель темы, настройки уведомлений.

## Props

| Prop | Тип | Обязательный | Описание |
|------|-----|--------------|----------|
| `user` | `Object` | да | Текущий пользователь |
| `isDarkTheme` | `Boolean` | да | Состояние тёмной темы |

## Emits

| Событие | Когда |
|---------|-------|
| `close` | Клик на оверлей или кнопку закрытия |
| `toggle-theme` | Клик на переключатель темы |
| `logout` | Клик на «Выйти из аккаунта» |
| `user-updated` | Успешное обновление профиля; данные: обновлённый `User` |

## Inject

- `showError`, `showInfo`

## Ключевое состояние (refs)

| Ref | Описание |
|-----|----------|
| `isEditProfileOpen` | Открыта ли секция редактирования |
| `editUsername` | Новое имя пользователя |
| `isChangePasswordOpen` | Открыта ли секция смены пароля |
| `oldPassword`, `newPassword`, `confirmPassword` | Поля формы смены пароля |
| `isNotificationsOpen` | Открыта ли секция настроек уведомлений |
| `emailConsents`, `webPushConsents` | Битовые маски согласий на уведомления |
| `isAvatarMenuOpen` | Открыто ли контекстное меню аватара |
| `avatarFileInput` | Ref на скрытый `<input type="file">` для загрузки аватара |

## Ключевые функции

| Функция | Описание |
|---------|----------|
| `loadSettings()` | Загружает настройки при монтировании |
| `toggleEmailConsent(bit)` | Переключает бит согласия email, сохраняет через `UpdateSettings` |
| `toggleWebPushConsent(bit)` | Переключает бит согласия web-push, сохраняет через `UpdateSettings` |
| `changePassword()` | Валидирует и вызывает `ChangePassword(dto)` |
| `updateUser()` | Вызывает `UpdateUser(dto)`, эмитит `user-updated` |
| `toggleAvatarMenu()` | Показывает/скрывает контекстное меню аватара |
| `triggerFileInput()` | Открывает диалог выбора файла для загрузки аватара |
| `uploadAvatar()` | Читает выбранный файл, вызывает `UpdateAvatar(byteArray)`, эмитит `user-updated` |
| `deleteAvatar()` | Вызывает `DeleteAvatar()`, эмитит `user-updated` с `avatarPath: null` |
| `formatDate(dateStr)` | Форматирует дату в `ru-RU` |

## Wails-биндинги

- `auth/Handler`: `ChangePassword`
- `users/Handler`: `UpdateUser`, `UpdateAvatar`, `DeleteAvatar`
- `settings/Handler`: `GetSettings`, `UpdateSettings`
