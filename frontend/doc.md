# frontend

## Назначение

Vue 3 фронтенд десктопного приложения KFC Chat. Собирается через Vite и встраивается в бинарник Wails через `//go:embed`.

## Стек

- **Vue 3** (`^3.3.4`) — Composition API через Options API с `setup()`
- **Vite** (`^4.4.5`) — сборка и dev-сервер (порт `5173`)
- **Без роутера, без стора** — состояние управляется через `ref`/`computed`, передаётся через `provide/inject` и `emit`

## Структура

```
frontend/
  index.html            — точка входа, подключает Wails IPC/runtime
  vite.config.js        — конфиг Vite, externals для Wails runtime
  package.json          — зависимости (только vue + vite)
  src/
    main.js             — создание и монтирование Vue-приложения
    App.vue / App.js    — корневой компонент, управление экранами
    components/         — все Vue-компоненты (по папкам)
    constants/          — константы приложения
    styles/             — глобальные CSS и дизайн-токены
    utils/              — утилиты (debounce)
  wailsjs/              — автогенерируемые Wails-биндинги (Go → JS)
```

## Сборка

```bash
# Dev-режим (запускается через wails dev)
npm run dev

# Продакшен-сборка (запускается через wails build)
npm run build
```

## Особенности

- Wails runtime (`/wails/runtime.js`, `/wails/ipc.js`) подключается как external в Vite — инъектируется Wails-фреймворком в рантайме.
- Биндинги в `wailsjs/go/` автогенерируются из Go-хендлеров при каждой сборке.
