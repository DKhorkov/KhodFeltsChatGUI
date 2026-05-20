# Пакет internal/common

## Назначение

Общие утилиты и константы, используемые по всему приложению: работа с датами и таймзонами, HTTP-заголовки, логирование, управление директорией данных приложения.

## Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| `DateFormat` | `"02.01.2006"` | Формат даты (дд.мм.гггг) |
| `DateTimeFormat` | `"02.01.2006 15:04:05"` | Формат даты и времени |
| `ContentTypeHeaderName` | `"Content-Type"` | Имя HTTP-заголовка Content-Type |
| `ApplicationJSONContentType` | `"application/json"` | Значение Content-Type для JSON |
| `CookieHeaderName` | `"Cookie"` | Имя HTTP-заголовка Cookie |
| `LoggingTraceSkipLevel` | `1` | Уровень пропуска стека для логгера |

## Переменные

| Переменная | Тип | Описание |
|------------|-----|----------|
| `Timezone` | `*time.Location` | Временная зона пользователя, определяется при запуске |
| `LogsDir` | `string` | Путь к директории логов (`<AppDataDir>/logs`) |
| `LogsPath` | `string` | Шаблон пути к файлу лога (`<LogsDir>/%s.log`) |

## Функции

| Функция | Описание |
|---------|----------|
| `AppDataDir() string` | Возвращает путь к директории данных приложения (OS-зависимый: `~/Library/Application Support/KhodFeltsChatGUI` на macOS, `~/.config/KhodFeltsChatGUI` на Linux, `%AppData%/KhodFeltsChatGUI` на Windows) |
| `CreateAppDataDir()` | Создаёт директорию данных приложения, если не существует |
| `CreateLogsDir()` | Создаёт директорию логов, если не существует |

## Зависимости

- `os`, `path/filepath`, `time`
- `time/tzdata` (для поддержки таймзон в Docker-контейнере)
