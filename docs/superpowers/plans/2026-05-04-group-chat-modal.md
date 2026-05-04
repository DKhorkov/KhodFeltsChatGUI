# Group Chat Modal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a modal window for group chats showing chat info and a clickable member list, with entry points from the sidebar avatar and conversation header title.

**Architecture:** New `GroupChatModal` component (vue/js/css triplet following existing pattern). ChatView gets a new `selectedGroupChat` ref and updated click handlers. Existing global CSS classes (`modal-overlay`, `modal-content`, `user-item`) are reused; new `group-modal__` classes handle group-specific styling.

**Tech Stack:** Vue 3 (Options API with `setup()`), CSS custom properties from existing design system.

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `frontend/src/components/GroupChatModal/GroupChatModal.vue` | Template: modal overlay, header (avatar + title + description), member list |
| Create | `frontend/src/components/GroupChatModal/GroupChatModal.js` | Logic: props, emits, formatDate helper |
| Create | `frontend/src/components/GroupChatModal/GroupChatModal.css` | Group-modal-specific styles |
| Modify | `frontend/src/components/ChatView/ChatView.js` | New ref `selectedGroupChat`, refactored `openMemberProfile`, new `openChatInfo`, new `openGroupMemberProfile` |
| Modify | `frontend/src/components/ChatView/ChatView.vue` | Import GroupChatModal, clickable avatar for all chats, clickable header title, render GroupChatModal |
| Modify | `frontend/src/components/ChatView/ChatView.css` | Clickable title styles in conversation header |

---

### Task 1: Create GroupChatModal component

**Files:**
- Create: `frontend/src/components/GroupChatModal/GroupChatModal.js`
- Create: `frontend/src/components/GroupChatModal/GroupChatModal.vue`
- Create: `frontend/src/components/GroupChatModal/GroupChatModal.css`

- [ ] **Step 1: Create GroupChatModal.js**

```js
import {computed} from 'vue'

export default {
    name: 'GroupChatModal',
    props: {
        chat: {
            type: Object,
            required: true,
        },
        currentUser: {
            type: Object,
            required: true,
        },
    },
    emits: ['close', 'open-member-profile'],

    setup(props, {emit}) {
        const chatTitle = computed(() => {
            return props.chat.title || `Чат #${props.chat.id}`
        })

        const chatInitial = computed(() => {
            return chatTitle.value.charAt(0).toUpperCase()
        })

        const membersCount = computed(() => {
            return props.chat.members ? props.chat.members.length : 0
        })

        const openMemberProfile = (member) => {
            emit('open-member-profile', member)
        }

        return {
            chatTitle,
            chatInitial,
            membersCount,
            openMemberProfile,
        }
    },
}
```

- [ ] **Step 2: Create GroupChatModal.vue**

```vue
<template>
  <div class="modal-overlay"
       @click="$emit('close')"
       @keydown.escape="$emit('close')"
       tabindex="-1"
       v-focus
  >
    <div class="modal-content group-modal" @click.stop>
      <button class="modal-content__close" @click="$emit('close')" title="Закрыть">&times;</button>

      <div class="group-modal__header">
        <div class="group-modal__avatar">
          {{ chatInitial }}
        </div>
        <div class="group-modal__title">
          <h2>{{ chatTitle }}</h2>
          <span v-if="chat.description" class="group-modal__description">{{ chat.description }}</span>
        </div>
      </div>

      <div class="group-modal__members">
        <h4 class="group-modal__members-heading">Участники ({{ membersCount }})</h4>
        <div class="modal-content__users-list">
          <div
              v-for="member in chat.members"
              :key="member.id"
              class="user-item user-item--clickable"
              @click="openMemberProfile(member)"
          >
            <div class="user-item__avatar">
              {{ member.username.charAt(0).toUpperCase() }}
            </div>
            <div class="user-item__info">
              <div class="user-item__name">
                {{ member.username }}
                <span v-if="member.id === currentUser?.id" class="group-modal__you-badge">(вы)</span>
              </div>
              <div v-if="member.email" class="user-item__email">{{ member.email }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script src="./GroupChatModal.js"></script>
<style src="./GroupChatModal.css"></style>
```

- [ ] **Step 3: Create GroupChatModal.css**

```css
.group-modal {
    width: 420px;
}

.group-modal__header {
    display: flex;
    align-items: center;
    gap: var(--space-lg);
    margin-bottom: var(--space-2xl);
}

.group-modal__avatar {
    width: 56px;
    height: 56px;
    background: var(--accent);
    color: var(--text-on-accent);
    border-radius: var(--radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: var(--font-xl);
    flex-shrink: 0;
}

.group-modal__title h2 {
    font-size: var(--font-xl);
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.group-modal__description {
    font-size: var(--font-sm);
    color: var(--text-muted);
    margin-top: var(--space-xs);
    display: block;
}

.group-modal__members {
    margin-top: var(--space-lg);
}

.group-modal__members-heading {
    font-size: var(--font-md);
    font-weight: 600;
    color: var(--text-secondary);
    margin: 0 0 var(--space-md) 0;
}

.group-modal__you-badge {
    font-size: var(--font-xs);
    color: var(--text-muted);
    font-weight: 400;
}

.user-item--clickable {
    cursor: pointer;
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/GroupChatModal/GroupChatModal.js frontend/src/components/GroupChatModal/GroupChatModal.vue frontend/src/components/GroupChatModal/GroupChatModal.css
git commit -m "feat: add GroupChatModal component"
```

---

### Task 2: Wire GroupChatModal into ChatView

**Files:**
- Modify: `frontend/src/components/ChatView/ChatView.js`
- Modify: `frontend/src/components/ChatView/ChatView.vue`
- Modify: `frontend/src/components/ChatView/ChatView.css`

- [ ] **Step 1: Update ChatView.js — add refs, refactor handlers, import component**

Add import at top of file (after EmojiPicker import):

```js
import GroupChatModal from '../GroupChatModal/GroupChatModal.vue'
```

Add to `components`:

```js
components: {EmojiPicker, GroupChatModal},
```

Add new ref after `selectedMember`:

```js
const selectedGroupChat = ref(null)
```

Replace the existing `openMemberProfile` function:

```js
const openMemberProfile = (chat) => {
    if (chat.type === CHAT_TYPE.PRIVATE) {
        const member = getOtherMember(chat)
        if (member) {
            selectedMember.value = member
        }
    } else {
        selectedGroupChat.value = chat
    }
}
```

Add new `openChatInfo` function (after `openMemberProfile`):

```js
const openChatInfo = () => {
    if (!selectedChat.value) return
    openMemberProfile(selectedChat.value)
}
```

Add new `openGroupMemberProfile` function (after `openChatInfo`):

```js
const openGroupMemberProfile = (member) => {
    selectedGroupChat.value = null
    selectedMember.value = member
}
```

Add to the return block:

```js
selectedGroupChat,
openChatInfo,
openGroupMemberProfile,
```

- [ ] **Step 2: Update ChatView.vue — make sidebar avatar clickable for all chats**

Replace line 34 (the `:class` on avatar):

Old:
```html
:class="{ 'chat-item__avatar--clickable': getOtherMember(chat) }"
```

New:
```html
:class="{ 'chat-item__avatar--clickable': true }"
```

- [ ] **Step 3: Update ChatView.vue — make conversation header title clickable**

Replace line 52:

Old:
```html
<h3>{{ getChatTitle(selectedChat) }}</h3>
```

New:
```html
<h3 class="conversation__header-title" @click="openChatInfo">{{ getChatTitle(selectedChat) }}</h3>
```

- [ ] **Step 4: Update ChatView.vue — add GroupChatModal render block**

Add after the existing member profile modal block (after line 145, before the closing `</div>` of `.chat-layout`):

```html
<!-- Информация о групповом чате -->
<GroupChatModal
    v-if="selectedGroupChat"
    :chat="selectedGroupChat"
    :current-user="currentUser"
    @close="selectedGroupChat = null"
    @open-member-profile="openGroupMemberProfile"
/>
```

- [ ] **Step 5: Update ChatView.css — add clickable title styles**

Add at the end of the conversation header section:

```css
.conversation__header-title {
    cursor: pointer;
    transition: color var(--transition-base);
}

.conversation__header-title:hover {
    color: var(--accent);
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ChatView/ChatView.js frontend/src/components/ChatView/ChatView.vue frontend/src/components/ChatView/ChatView.css
git commit -m "feat: wire GroupChatModal into ChatView with sidebar and header entry points"
```

---

### Task 3: Manual verification

- [ ] **Step 1: Start the dev server and verify**

Run the app and check:
1. Private chat — clicking avatar in sidebar still opens member profile
2. Private chat — clicking title in conversation header opens member profile
3. Group chat — clicking avatar in sidebar opens group modal with title, description, member list
4. Group chat — clicking title in conversation header opens group modal
5. Group modal — clicking a member closes group modal and opens their profile
6. Group modal — "(вы)" badge shown next to current user
7. All modals close on overlay click, X button, and Escape key

- [ ] **Step 2: Commit any fixes if needed**
