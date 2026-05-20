# Пакет internal/repositories/auth

## Назначение

HTTP-репозиторий для операций аутентификации и управления паролями. Реализует интерфейс `interfaces.AuthRepository`. Все методы потокобезопасны (`sync.Mutex`).

## Типы

| Тип | Описание |
|-----|----------|
| `Repository` | Встраивает `base.Repository`; содержит `httpClient`, `baseURL`, `mu sync.Mutex` |

## Методы

| Метод | HTTP | Эндпоинт | Описание |
|-------|------|----------|----------|
| `Register(ctx, registerData)` | POST | `/users` | Регистрация пользователя, возвращает `*domains.User` |
| `Login(ctx, loginData)` | POST | `/sessions` | Логин; извлекает `accessToken` и `refreshToken` из cookies ответа |
| `Logout(ctx, accessToken)` | DELETE | `/sessions` | Завершение сессии |
| `RefreshTokens(ctx, refreshToken)` | PUT | `/sessions` | Обновление пары токенов через refresh-cookie |
| `SendVerifyEmailMessage(ctx, email)` | POST | `/users/email/verify` | Отправка письма подтверждения email |
| `SendForgetPasswordMessage(ctx, email)` | POST | `/users/password/forget` | Отправка письма восстановления пароля |
| `ChangePassword(ctx, accessToken, data)` | POST | `/users/password/change` | Смена пароля авторизованным пользователем |
| `ForgetPassword(ctx, token, newPassword)` | POST | `/users/password/forget/{token}` | Сброс пароля по токену из письма |

## Константы

- `accessTokenCookieName` = `"accessToken"`
- `refreshTokenCookieName` = `"refreshToken"`

## Зависимости

- `internal/repositories/base` — базовый репозиторий (`CloseBody`)
- `internal/interfaces` — `HTTPClient`
- `internal/domains` — DTO (`RegisterDTO`, `LoginDTO`, `TokensDTO`, `ChangePasswordDTO`, `ForgetPasswordDTO`, `SendVerifyEmailMessageDTO`, `SendForgetPasswordMessageDTO`)
- `internal/errors` — `ErrLogin`, `ErrRefreshTokens`
- `internal/common` — HTTP-заголовки
