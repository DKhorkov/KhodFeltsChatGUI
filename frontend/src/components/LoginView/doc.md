# Компонент LoginView

## Назначение

Экран входа и регистрации. Содержит две вкладки: «Вход» и «Регистрация», ссылки для повторной отправки письма подтверждения и сброса пароля.

## Emits

| Событие | Когда |
|---------|-------|
| `login-success` | Успешный вход |
| `show-forget-password` | Отправка письма сброса пароля; данные: строка с сообщением |

## Inject

- `showError`, `showInfo`

## Ключевое состояние (refs)

| Ref | Описание |
|-----|----------|
| `activeTab` | Активная вкладка: `TAB.LOGIN` или `TAB.REGISTER` |
| `loginForm` | `{ login, password }` |
| `registerForm` | `{ email, username, password, confirmPassword }` |

## Ключевые функции

| Функция | Описание |
|---------|----------|
| `handleLogin()` | Валидирует поля, вызывает `Login(loginForm)`, эмитит `login-success` |
| `handleRegister()` | Валидирует поля, вызывает `Register(registerForm)`, переключает на вкладку входа |
| `sendVerifyEmail()` | Вызывает `SendVerifyEmail(login)`, показывает инфо |
| `sendForgetPassword()` | Вызывает `SendForgetPassword(login)`, эмитит `show-forget-password` |

## Wails-биндинги

`auth/Handler`: `Login`, `Register`, `SendVerifyEmail`, `SendForgetPassword`
