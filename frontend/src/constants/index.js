export const THEME = Object.freeze({
    LIGHT: 0,
    DARK: 1,
})

export const CHAT_TYPE = Object.freeze({
    PRIVATE: 'private',
    GROUP: 'group',
})

export const VIEW = Object.freeze({
    LOADING: 'loading',
    LOGIN: 'login',
    CHAT: 'chat',
})

export const TAB = Object.freeze({
    LOGIN: 'login',
    REGISTER: 'register',
})

export const WAILS_EVENT = Object.freeze({
    NEW_MESSAGE: 'new_message',
    MESSAGE_DELETED: 'message_deleted',
    MESSAGE_EDITED: 'message_edited',
    REACTION_ADDED: 'reaction_added',
    REACTION_REMOVED: 'reaction_removed',
    CHATS_UPDATED: 'chats_updated',
    OPEN_CHAT: 'open_chat',
})

export const MESSAGES_PAGE_SIZE = 10
export const NOTIFICATION_DURATION_MS = 3000
export const SEARCH_DEBOUNCE_MS = 500
export const EMOJI_CLOSE_DELAY_MS = 500
export const HIGHLIGHT_DURATION_MS = 1500