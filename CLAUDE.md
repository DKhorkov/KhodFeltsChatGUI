# KhodFelts Chat GUI

Wails v2 десктопное приложение для чата. Go-бэкенд + Vue 3 фронтенд.

## Быстрый старт

```bash
# Запуск в dev-режиме
wails dev

# Сборка
wails build
```

## Архитектура

- **Бэкенд:** Go, трёхслойная архитектура Handler -> UseCases -> Repositories
- **Фронтенд:** Vue 3 (Composition API), Vite, без роутера, без стора
- **IPC:** Wails bindings — Vue вызывает Go-функции напрямую через автогенерируемые JS-биндинги
- **Real-time:** WebSocket + Wails EventsEmit/EventsOn для новых сообщений и обновления чатов

## Документация

Перед тем как исследовать кодовую базу заново, используй существующую документацию:

### Бэкенд (`docs/backend/`)

- [architecture.md](docs/backend/architecture.md) — общая архитектура, точка входа, жизненный цикл приложения
- [config.md](docs/backend/config.md) — переменные окружения и конфигурация
- [domains.md](docs/backend/domains.md) — все доменные модели и DTO
- [interfaces.md](docs/backend/interfaces.md) — все интерфейсы с сигнатурами методов
- [handlers.md](docs/backend/handlers.md) — Wails-хендлеры (7 штук), их методы
- [usecases.md](docs/backend/usecases.md) — бизнес-логика
- [repositories.md](docs/backend/repositories.md) — слой данных (HTTP, WebSocket, локальное хранилище)
- [errors.md](docs/backend/errors.md) — типы ошибок и маппинг в пользовательские сообщения

### Фронтенд (`docs/frontend/`)

- [architecture.md](docs/frontend/architecture.md) — стек, паттерны состояния, структура компонентов
- [components.md](docs/frontend/components.md) — все компоненты: назначение, props, emits, ключевые функции
- [constants.md](docs/frontend/constants.md) — все константы приложения
- [wails-bindings.md](docs/frontend/wails-bindings.md) — автогенерируемые биндинги, runtime events, модели
- [styles.md](docs/frontend/styles.md) — дизайн-токены, темы, общие CSS-классы
- [provide_inject_emit.md](docs/frontend/provide_inject_emit.md) — паттерны передачи состояния

## Структура проекта

```
cmd/v2/main.go              — точка входа Wails v2
internal/
  v2/application/           — App (Startup/Shutdown/BindHandlers)
  v2/handlers/              — 7 хендлеров (auth, chat, create_chat, search_users, forget_password, profile, theme)
  config/                   — конфигурация из env
  domains/                  — доменные модели
  errors/                   — ошибки и маппер
  interfaces/               — все интерфейсы
  usecases/                 — бизнес-логика
  repositories/             — репозитории (auth, users, chats, tokens, settings, ws, base)
frontend/
  src/components/           — Vue-компоненты (каждый: папка с .vue + .js + .css)
  src/constants/            — константы
  src/styles/global.css     — глобальные стили и дизайн-токены
  wailsjs/                  — автогенерируемые Wails-биндинги
```

## Соглашения

- Язык интерфейса и комментариев в коде: русский
- Каждый Vue-компонент — папка с тремя файлами: `Component.vue`, `Component.js`, `Component.css`
- Options API с `setup()` (не `<script setup>`)
- CSS: BEM-подобные имена классов с `__` и `--`
- Go: интерфейсы определены в `internal/interfaces/`, реализации — в соответствующих пакетах
- Моки генерируются через `go:generate mockgen` в `mocks/`
- Линтер: `.golangci.yaml`

## Обязательные правила

- **Правило:** при изменении Go-кода **обязательно** напиши или обнови тесты для затронутого кода.
- **Правило:** при изменении кода в директории **обязательно** обнови `doc.md` в этой же директории, чтобы документация соответствовала актуальному состоянию кода.
