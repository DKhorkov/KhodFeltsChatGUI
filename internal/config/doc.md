# Пакет internal/config

## Назначение

Конфигурация приложения, загружаемая из переменных окружения с дефолтными значениями.

## Типы

### `Config`

Корневая структура конфигурации.

| Поле | Тип | Описание |
|------|-----|----------|
| `HTTP` | `HTTPConfig` | Настройки HTTP-клиента и WebSocket |
| `Logging` | `logging.Config` | Настройки логирования |
| `Validation` | `ValidationConfig` | Регулярные выражения для валидации |

### `HTTPConfig`

| Поле | Тип | Env-переменная | Значение по умолчанию |
|------|-----|----------------|----------------------|
| `Timeout` | `time.Duration` | `HTTP_CLIENT_TIMEOUT` | `5s` |
| `WebsocketURL` | `string` | `HTTP_WEBSOCKET_URL` | `wss://kfc.webtm.ru/api` |
| `BaseURL` | `string` | `HTTP_BASE_URL` | `https://kfc.webtm.ru/api` |

### `ValidationConfig`

| Поле | Тип | Env-переменная | Описание |
|------|-----|----------------|----------|
| `EmailRegExp` | `string` | `EMAIL_REGEXP` | Регулярное выражение для email |
| `PasswordRegExps` | `[]string` | `PASSWORD_REGEXPS` | Набор правил: длина >= 8, строчная буква, заглавная буква, цифра, спецсимвол |
| `UsernameRegExps` | `[]string` | `USERNAME_REGEXPS` | Набор правил: длина 5-70, только латиница и цифры |

## Функции

| Функция | Описание |
|---------|----------|
| `New() Config` | Создаёт конфигурацию, загружая значения из env-переменных |

## Зависимости

- `github.com/DKhorkov/kfcGUI/internal/common` — пути логов, формат даты, таймзона
- `github.com/DKhorkov/libs/loadenv` — загрузка env-переменных
- `github.com/DKhorkov/libs/logging` — конфиг логирования
