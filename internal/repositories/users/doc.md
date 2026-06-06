# Пакет internal/repositories/users

## Назначение

HTTP-репозиторий для работы с пользователями. Реализует интерфейс `interfaces.UsersRepository`. Потокобезопасен: чтение через `RLock`, запись через `Lock` (`sync.RWMutex`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.RWMutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `GetCurrentUser(ctx, accessToken)` | GET | `/users/me` | Получение текущего пользователя по access-токену |
| `UpdateUser(ctx, accessToken, updateUserData)` | PUT | `/users/me` | Обновление профиля текущего пользователя |
| `SearchUsers(ctx, filters, pagination)` | GET | `/users` | Поиск пользователей с фильтром `username` и пагинацией; не требует авторизации |
| `UpdateAvatar(ctx, accessToken, fileData)` | PUT | `/users/me/avatar` | Загрузка/обновление аватара (multipart/form-data); возвращает URL аватара |
| `DeleteAvatar(ctx, accessToken)` | DELETE | `/users/me/avatar` | Удаление аватара; ожидает 204 No Content |

## Константы

- `accessTokenCookieName` = `"accessToken"`
- `limitQueryParamName` = `"limit"`
- `offsetQueryParamName` = `"offset"`
- `avatarFormFileKey` = `"avatar"` — ключ поля файла в multipart-форме
- `avatarFileName` = `"avatar.jpg"` — имя файла аватара в multipart-запросе

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — `User`, `UpdateUserDTO`, `UsersFilters`, `Pagination`
- `internal/common` — HTTP-заголовки
