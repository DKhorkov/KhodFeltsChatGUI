# Group Chat Modal — Design Spec

## Summary

Add a modal window for group chats that shows chat info (name, description) and a list of members. Clicking a member opens their profile using the existing member profile modal.

## Points of Entry

1. **Avatar in sidebar** — clicking the avatar of a group chat opens the group modal (currently only works for private chats)
2. **Chat title in conversation header** — clicking the `<h3>` title opens the modal. For private chats, opens the member profile; for group chats, opens the group modal.

## New Component: `GroupChatModal`

Location: `frontend/src/components/GroupChatModal/`

Files:
- `GroupChatModal.vue` — template
- `GroupChatModal.js` — logic
- `GroupChatModal.css` — styles (or rely on global.css like ProfileModal)

### Props
- `chat` (Object, required) — the group chat object with `title`, `description`, `members`
- `currentUser` (Object, required) — to identify the current user in the members list

### Emits
- `close` — close the modal
- `open-member-profile(member)` — request to open a member's profile

### Template Structure

```
modal-overlay
  modal-content group-modal
    close button (x)
    header:
      avatar (first letter of chat title)
      title (chat.title or "Chat #id")
      description (chat.description, if present)
    members section:
      heading "Participants (N)"
      list of members:
        each member: avatar (first letter) + username
        clickable -> emits open-member-profile
```

## Changes to ChatView

### ChatView.js
- New `ref`: `selectedGroupChat` (null or chat object)
- Refactor `openMemberProfile(chat)`:
  - Private chat: sets `selectedMember` (existing behavior)
  - Group chat: sets `selectedGroupChat`
- New function `openChatInfo(chat)` — called from conversation header title click. Same logic as refactored `openMemberProfile`.
- Handler for `open-member-profile` event from GroupChatModal: sets `selectedGroupChat = null`, then sets `selectedMember = member`
- Export new refs/functions in return block

### ChatView.vue
- Import and register `GroupChatModal` component
- Sidebar avatar: make clickable for all chat types (remove condition that limits `--clickable` class to private chats only)
- Conversation header `<h3>`: wrap in clickable element, call `openChatInfo(selectedChat)`
- Add `<GroupChatModal>` render block conditioned on `selectedGroupChat`

### ChatView.css
- Clickable title styles (cursor, hover effect)

## Styling

- Reuse existing `modal-overlay`, `modal-content`, `modal-content__close` classes
- New `group-modal__` prefixed classes for group-specific elements
- Member list items: hover effect, pointer cursor, consistent with existing chat-item styling
- Avatar circles: same style as sidebar avatars (accent background, white text, first letter)

## Out of Scope

- Editing chat name/description
- Adding/removing members
- Refactoring the inline member profile modal in ChatView.vue (lines 111-145)
