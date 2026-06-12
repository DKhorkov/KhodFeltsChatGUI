# GUI Badge Counter — Design

## Цель

Поддержать в десктопном GUI ту же фичу непрочитанных сообщений, что недавно сделана в PWA (`KhodFeltsChat`, ветка `counters`, см. [PWA-design](../../../../../KhodFeltsChat/docs/superpowers/specs/2026-06-11-pwa-badge-counter-design.md)):

1. **Число непрочитанных** на каждом чате в списке (вместо абстрактной точки `chat-item__unread-dot`).
2. **Системные OS-уведомления** через готовый каркас `wailsruntime.SendNotification`, но с корректным правилом подавления.
3. **Счётчик в заголовке окна** (Dock/taskbar overlay не используем; `wailsruntime.WindowSetTitle` — кросс-платформенно).
4. **Открытие нужного чата при клике** по системному уведомлению (полностью уже плюмбится через событие `open_chat`, нужно убедиться что работает).

## Проблема

Сейчас в GUI:

- `domains.Chat` имеет бинарный `IsRead bool`, число непрочитанных нигде не считается. На чатах рисуется маленький кружок `.chat-item__unread-dot` — пользователь не видит, сколько именно сообщений ждёт его в каждом чате и в сумме.
- В `ChatView.js:200` уже есть вызов `notifications.ShowNotification(...)`, но он гейтится только по `!isWindowFocused` — если пользователь в открытом приложении читает другой чат, новое сообщение в любом другом чате не вызывает OS-уведомления. Это неудобно: входящее теряется до 5 сек polling'а / следующего ручного `loadChats`.
- Заголовок окна (`KFC Chat`) статичен — нет ни одного пассивного сигнала, что есть что-то новое, когда окно свёрнуто или скрыто за другими.
- Бэк уже отдаёт `Chat.UnreadCount` (ветка `counters` в `KhodFeltsChat`) — поле просто игнорируется в GUI.

## Принципы решения

**Единственный источник истины — БД, агрегированная в `GET /chats`.** Никаких локальных «было N стало N+1». Каждый раз клиент пересчитывает абсолютное число из свежего списка чатов. Любая рассинхронизация (потерянный WS-кадр, прочтение на другом устройстве, удаление «у всех») лечится одним очередным `loadChats` — а он и так дёргается на каждое значимое событие + раз в 5 сек.

**`loadChats` — единственная точка обновления бейджа.** Логика: после успешного `GetUserChats` считается `total = Σ chat.unreadCount` и пишется в title окна через новый Wails-binding `SetUnreadBadge(total)`. То же место, где уже рендерится список чатов с per-chat бейджами.

**Подавление OS-уведомления = окно в фокусе И целевой чат открыт.** Любая другая комбинация (окно НЕ в фокусе ИЛИ открыт другой чат) → показываем системное уведомление. Гейтится дополнительно `webPushConsents.ConsentNewMessage` — единый тоггл уведомлений (уже существует в настройках, используется и PWA, и GUI).

**Заголовок окна вместо нативного бейджа.** Wails v2.12.0 не имеет API для Dock badge / taskbar overlay. Реализация через cgo/syscall — отдельный проект на каждую ОС. Префикс `(N)` в title окна — кросс-платформенный сигнал, видный и в свёрнутом окне (в Dock-иконке macOS отображается название окна при наведении, на Windows — в taskbar text), и в alt-tab переключателе.

## Изменения по слоям GUI

### `internal/domains/chat.go`

Добавить поле `UnreadCount`. Поле `IsRead` (бинарный флаг) удалить — его семантика полностью покрывается `UnreadCount > 0`.

```go
type Chat struct {
    ID          uint64    `json:"id"`
    Title       *string   `json:"title,omitempty"`
    Description *string   `json:"description,omitempty"`
    Type        ChatType  `json:"type"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    UnreadCount uint64    `json:"unreadCount"`
    Members     []User    `json:"members,omitempty"`
    Messages    []Message `json:"messages,omitempty"`
}
```

`chat_test.go` — добавить `UnreadCount` в существующие фикстуры. Места во фронте и в `internal/v1/windows/chat/window.go`, где раньше использовался `chat.IsRead`, переключаются на `chat.unreadCount > 0` (для bold-заголовка, классов и индикаторов).

### `internal/repositories/chats/repository.go`

**Без изменений.** Поле `unreadCount` приходит в JSON-ответе `/chats` и автоматически десериализуется в `domains.Chat`. Тесты репозитория обновить — добавить поле в моковые ответы и в проверки.

### `internal/v2/handlers/notifications/handler.go`

Добавить метод `SetUnreadBadge`:

```go
const appTitle = "KFC Chat" // совпадает с options.App.Title в cmd/v2/main.go

func (h *Handler) SetUnreadBadge(total int) {
    var title string
    switch {
    case total <= 0:
        title = appTitle
    case total > 99:
        title = "(99+) " + appTitle
    default:
        title = fmt.Sprintf("(%d) %s", total, appTitle)
    }
    wailsruntime.WindowSetTitle(h.wailsCtx, title)
}
```

Метод биндится через `BindHandlers` в `internal/v2/application/`, чтобы появилось `wailsjs/go/notifications/Handler.js#SetUnreadBadge`. Метод выбран в notifications-хендлере, потому что он уже владеет OS-уровневыми сайд-эффектами (SendNotification, focusWindow). Альтернатива — выделить badge-handler — overkill для одного метода.

Тесты хендлера (`handler_test.go`):
- `SetUnreadBadge(0)` — `WindowSetTitle("KFC Chat")`
- `SetUnreadBadge(7)` — `WindowSetTitle("(7) KFC Chat")`
- `SetUnreadBadge(100)` — `WindowSetTitle("(99+) KFC Chat")`
- `SetUnreadBadge(-1)` — `WindowSetTitle("KFC Chat")` (защита от мусора)

Поскольку `WindowSetTitle` — глобальная функция из Wails runtime, тест-двойник делается через вспомогательную инжектируемую функцию или интерфейс `windowAdapter` (по тому же паттерну, как остальные тесты в этом пакете).

### `internal/v2/application/app.go` (или где BindHandlers)

Зарегистрировать `SetUnreadBadge` в списке биндингов notifications-хендлера. Если хендлер уже целиком биндится — изменений нет, новый метод подхватится автогенерацией Wails CLI.

### `frontend/src/components/ChatView/ChatView.js`

**1) Импорт нового биндинга:**

```js
import {SetUnreadBadge} from '../../../wailsjs/go/notifications/Handler'
```

**2) Функция расчёта и публикации total:**

```js
const updateUnreadBadge = (chatsList) => {
    const total = chatsList.reduce((sum, c) => sum + (c.unreadCount || 0), 0)
    SetUnreadBadge(total).catch(err => console.error('Не удалось обновить бейдж:', err))
}
```

**3) Точки вызова:**

- В `loadChats()` после успешного `GetUserChats` — перед `return chats`.
- В обработчике `CHATS_UPDATED` (когда Go-side polling 5 сек прислал свежий список): `handleChatsUpdated(updatedChats) { chats.value = updatedChats; updateUnreadBadge(updatedChats) }`.
- В `App.js#handleLogout` — `SetUnreadBadge(0)` перед чисткой состояния (чтобы при выходе и возврате на login-скрин в title не висело старое значение).

Других точек не нужно: WS new/delete/edit и так зовут `loadChats`, отправка сообщения — тоже, открытие чата — тоже (там бэк помечает прочитанным и `loadChats` возвращает уменьшенный `unreadCount`).

**4) Новое условие подавления уведомления (`handleNewMessage`):**

```js
const isThisChatActive = isWindowFocused && selectedChat.value?.id === message.chatId
const consentGiven = (webPushConsents.value & CONSENT_NEW_MESSAGE) !== 0

if (!isThisChatActive && consentGiven) {
    ShowNotification(message.sender.username, message.text, message.chatId)
        .catch(err => console.error('Ошибка системного уведомления:', err))
}
```

Сводная таблица поведения:

| Окно в фокусе | Активный чат = чат с новым сообщением | OS-уведомление |
|---|---|---|
| Да | Да | НЕТ (подавляем) |
| Да | Нет (другой чат / без выбранного) | ДА |
| Нет | Да | ДА |
| Нет | Нет | ДА |

Дополнительно гейтится `webPushConsents.ConsentNewMessage` — если пользователь выключил уведомления, OS-нотификации не показываем независимо от состояния окна.

### `frontend/src/components/ChatView/ChatView.vue`

Заменить рендер `.chat-item__unread-dot` на `.chat-item__unread-badge` с числом:

```vue
<div
    v-for="chat in chats"
    :key="chat.id"
    :class="['chat-item',
             { 'chat-item--active': selectedChat?.id === chat.id,
               'chat-item--unread': chat.unreadCount > 0 }]"
    @click="selectChat(chat)"
>
    <!-- ... аватар, заголовок, превью ... -->
    <div v-if="chat.unreadCount > 0" class="chat-item__unread-badge">
        {{ chat.unreadCount > 99 ? '99+' : chat.unreadCount }}
    </div>
</div>
```

Класс `chat-item--unread` теперь определяется по `unreadCount > 0` (а не `!chat.isRead`). Поведение идентично, но устраняем зависимость от поля, которое инвариантно с `unreadCount`.

### `frontend/src/components/ChatView/ChatView.css`

Заменить `.chat-item__unread-dot` (точка 8×8) на `.chat-item__unread-badge` (пилюля с числом). Стилистика совпадает с PWA (для единообразия):

```css
.chat-item__unread-badge {
    min-width: 20px;
    height: 20px;
    padding: 0 6px;
    background: var(--accent);
    color: #fff;
    border-radius: var(--radius-full);
    font-size: var(--font-xs);
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    margin-left: var(--space-sm);
    line-height: 1;
}
```

Старый класс `.chat-item__unread-dot` удалить.

## Поведение

| Сценарий | Бейдж в списке чатов | Title окна | OS-уведомление |
|---|---|---|---|
| Приходит сообщение в другой чат, окно в фокусе | +1 в том чате при ближайшем `loadChats` | пересчитывается | показывается (consent ON) |
| Приходит сообщение в активный чат, окно в фокусе | per-chat бейдж может коротко мигнуть в списке (бэк mark-as-read срабатывает при следующем `loadMessages` / открытии чата) | временно увеличивается на 1, синхронизируется при следующем mark-as-read | подавлено |
| Приходит сообщение, окно НЕ в фокусе | +1 | `(N+1)` → `(N+2)` … | показывается (consent ON) |
| Пользователь открыл чат с непрочитанными | бейдж этого чата исчезает (бэк mark-as-read через `GET /chats/{id}/messages`) | `(N)` → `(N − read)` | — |
| Сообщение удалили «у всех», было непрочитанным | бейдж этого чата уменьшается на 1 | total уменьшается на 1 | — |
| Прочитал на другом устройстве | при следующем polling (≤5 сек) или WS-событии — синхронизируется | то же | — |
| Logout | — | сбрасывается на `KFC Chat` | — |
| WS отвалился | синхронизируется при восстановлении / на polling | то же | — |

## Сценарий клика по уведомлению

Уже плюмбится (нужно только проверить, что работает на практике):

1. `SendNotification` пишет `chatId` в `userInfo`.
2. `OnNotificationResponse` (в `notifications.Handler.SetContext`) при клике:
   - `focusWindow()` — `WindowUnminimise` + `WindowShow` + `SetAlwaysOnTop(true→false)`.
   - `EventsEmit("open_chat", chatID)`.
3. `ChatView.js:503` слушает `OPEN_CHAT` и вызывает `openChatById(chatId)`.

В рамках этой ветки — smoke-тест ручной: запустить `wails dev`, свернуть окно, отправить сообщение с другого аккаунта, кликнуть по нотификации, проверить что окно поднялось и нужный чат открылся.

## Что НЕ делается в этой ветке

- Нативный бейдж на Dock (macOS) / overlay icon на taskbar (Windows). Если потребуется — отдельная ветка с cgo/syscall.
- Звук уведомления — оставляем дефолт ОС (берётся из `wailsruntime.SendNotification`).
- Группировка уведомлений из одного чата.
- Отдельный desktop-тоггл уведомлений в настройках. Используем существующий `webPushConsents.ConsentNewMessage` — это удобно: один свитч управляет и PWA, и десктопом.
- Новые endpoint'ы на бэке. Всё уже отдаётся в `GET /chats`.
- Изменения в `Message` домене — `isRead` сообщения остаётся как был (это другая семантика — статус конкретного сообщения, а не чата).

## Совместимость

- `Chat.UnreadCount` — добавление поля, GUI начинает читать то, что бэк уже отдаёт.
- `Chat.IsRead` — удаление поля. Все потребители (`v1/windows/chat`, `frontend/src/components/ChatView`) переключаются на `unreadCount > 0`. Бэк всё ещё отдаёт `isRead` в JSON, но GUI его игнорирует.
- Новый Wails-метод `SetUnreadBadge` — чисто аддитивный, ничего не ломает.

## Тестирование

| Тест | Что проверяем |
|---|---|
| `internal/domains/chat_test.go` | Существующие сценарии не ломаются при добавлении нового поля. Фикстуры пополняются `UnreadCount`. |
| `internal/repositories/chats/repository_test.go` | `unreadCount` корректно десериализуется из JSON ответа `/chats` (мокаем HTTP). |
| `internal/v2/handlers/notifications/handler_test.go` | `SetUnreadBadge` — четыре кейса: 0, обычное число, > 99, отрицательное. Проверяем форматирование переданного в `WindowSetTitle` значения. |
| Ручной smoke-тест | (1) Чат-бейдж рендерится с правильным числом и `99+`. (2) Title окна меняется на `(N) KFC Chat`. (3) Подавление работает по таблице из «Поведение». (4) Клик по нотификации поднимает окно и открывает нужный чат. (5) Logout сбрасывает title. |

## Обновление doc.md

По соглашению проекта (`CLAUDE.md`) — при изменении кода в директории обновить соответствующий `doc.md`:

- `internal/domains/doc.md` — пополнить описание `Chat` полем `UnreadCount`.
- `internal/v2/handlers/notifications/doc.md` — добавить метод `SetUnreadBadge`.
- `frontend/src/components/ChatView/doc.md` — описать `updateUnreadBadge`, новое условие подавления, новый класс `.chat-item__unread-badge`.

## Открытые вопросы

Нет.
