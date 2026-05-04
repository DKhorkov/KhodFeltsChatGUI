# Конфигурация

Файл: `internal/config/config.go`

## Структура конфигурации

```go
type Config struct {
    HTTP       HTTPConfig
    Logging    logging.Config
    Validation ValidationConfig
}
```

### HTTPConfig

```go
type HTTPConfig struct {
    Timeout      time.Duration
    WebsocketURL string
    BaseURL      string
}
```

### ValidationConfig

```go
type ValidationConfig struct {
    EmailRegExp     string
    PasswordRegExps []string
    UsernameRegExps []string
}
```

`PasswordRegExps` и `UsernameRegExps` — срезы, а не одно регулярное выражение, потому что Go-реализация regexp не поддерживает lookahead/backtracking. Каждое правило проверяется отдельно; значение считается корректным, если оно удовлетворяет всем правилам.

## Переменные окружения

Конфиг читается функцией `config.New()` при старте. Переменные загружаются из `.env`-файла через `loadenv.Init()`.

| Переменная | Тип | Значение по умолчанию | Описание |
|---|---|---|---|
| `HTTP_CLIENT_TIMEOUT` | int (секунды) | `5` | Таймаут HTTP-клиента |
| `HTTP_BASE_URL` | string | `http://185.119.59.215:8080` | Базовый URL REST API |
| `HTTP_WEBSOCKET_URL` | string | `ws://185.119.59.215:8080` | URL для WebSocket |
| `EMAIL_REGEXP` | string | `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$` | Регулярное выражение для email |
| `PASSWORD_REGEXPS` | string (разделитель `;`) | см. ниже | Правила валидации пароля |
| `USERNAME_REGEXPS` | string (разделитель `;`) | см. ниже | Правила валидации логина |

### Правила пароля по умолчанию

```
.{8,}       — минимум 8 символов
[a-z]       — хотя бы одна буква в нижнем регистре
[A-Z]       — хотя бы одна буква в верхнем регистре
[0-9]       — хотя бы одна цифра
[^\d\w]     — хотя бы один спецсимвол
```

### Правила username по умолчанию

```
^.{5,70}$       — длина от 5 до 70 символов
^[A-Za-z0-9]+$  — только латинские буквы и цифры
```

## Логирование

Уровень логирования всегда `DEBUG`. Путь к файлу лога формируется в момент запуска:

```go
fmt.Sprintf(common.LogsPath, time.Now().In(common.Timezone).Format(common.DateFormat))
```

Файл лога создаётся в директории данных приложения (см. `internal/common/appdata.go`):
- macOS: `~/Library/Application Support/KhodFeltsChatGUI/`
- Linux: `~/.config/KhodFeltsChatGUI/`
- Windows: `%AppData%/KhodFeltsChatGUI/`
