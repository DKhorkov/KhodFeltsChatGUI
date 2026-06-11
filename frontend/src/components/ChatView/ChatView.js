import {inject, nextTick, onMounted, onUnmounted, ref, watch} from 'vue'
import {
    GetUserChats,
    StartListening,
    StopListening
} from '../../../wailsjs/go/chats/Handler'
import {
    DeleteMessage,
    GetChatMessages,
    GetMessageByID,
    SendMessage,
    UpdateMessage
} from '../../../wailsjs/go/messages/Handler'
import {GetCurrentUser} from '../../../wailsjs/go/users/Handler'
import {GetTheme, ToggleTheme} from '../../../wailsjs/go/theme/Handler'
import {GetSettings} from '../../../wailsjs/go/settings/Handler'
import {ShowNotification} from '../../../wailsjs/go/notifications/Handler'
import {CHAT_TYPE, EMOJI_CLOSE_DELAY_MS, HIGHLIGHT_DURATION_MS, MESSAGES_PAGE_SIZE, THEME, WAILS_EVENT} from '../../constants'

const CONSENT_NEW_MESSAGE = 1
import EmojiPicker from '../EmojiPicker/EmojiPicker.vue'
import GroupChatModal from '../GroupChatModal/GroupChatModal.vue'

export default {
    name: 'ChatView',
    components: {EmojiPicker, GroupChatModal},
    emits: ['logout', 'show-create-chat', 'show-search-users', 'show-profile', 'new-message-notification'],

    setup(props, {emit}) {
        const showError = inject('showError')
        const chats = ref([])
        const selectedChat = ref(null)
        const messages = ref([])
        const currentUser = ref(null)
        const newMessage = ref('')
        const messagesListRef = ref(null)
        const textareaRef = ref(null)
        const isDarkTheme = ref(false)
        const isEmojiPickerVisible = ref(false)
        let emojiCloseTimer = null

        const showEmojiPicker = () => {
            clearTimeout(emojiCloseTimer)
            isEmojiPickerVisible.value = true
        }

        const scheduleEmojiClose = () => {
            emojiCloseTimer = setTimeout(() => {
                isEmojiPickerVisible.value = false
            }, EMOJI_CLOSE_DELAY_MS)
        }
        const selectedMember = ref(null)
        const avatarZoomSrc = ref(null)
        const selectedGroupChat = ref(null)
        const webPushConsents = ref(0)
        const replyToMessage = ref(null)
        const highlightedMessageId = ref(null)
        const editingMessage = ref(null)
        const contextMenu = ref({ visible: false, x: 0, y: 0, message: null, deleteExpanded: false })
        // unreadMessageIds — реактивный Set ID сообщений, прилетевших пока пользователь был отскроллен вверх.
        // Храним именно ID (а не число), чтобы при удалении сообщения «у всех» корректно вычесть его из badge.
        const unreadMessageIds = ref(new Set())
        const isAtBottom = ref(true)
        let lastMessageObserver = null

        let isLoadingMore = false
        let hasMoreMessages = true
        let isWindowFocused = true

        const onWindowFocus = () => { isWindowFocused = true }
        const onWindowBlur = () => { isWindowFocused = false }

        const reloadSettings = async () => {
            try {
                const settings = await GetSettings()
                webPushConsents.value = settings.webPushConsents

                const newIsDark = settings.theme === THEME.DARK
                if (isDarkTheme.value !== newIsDark) {
                    isDarkTheme.value = newIsDark
                    document.documentElement.setAttribute('data-bs-theme', newIsDark ? 'dark' : 'light')
                }
            } catch (err) {
                console.error('Ошибка загрузки настроек:', err)
            }
        }

        const loadChats = async () => {
            try {
                chats.value = await GetUserChats(null)
            } catch (err) {
                console.error("Ошибка загрузки чатов:", err)
            }
        }

        const loadMessages = async (chatId) => {
            try {
                const fetched = await GetChatMessages(chatId, {
                    limit: MESSAGES_PAGE_SIZE,
                    offset: 0,
                })
                messages.value = fetched.reverse()
                hasMoreMessages = fetched.length >= MESSAGES_PAGE_SIZE

                await nextTick()
                scrollToBottom()
            } catch (err) {
                console.error("Ошибка загрузки сообщений:", err)
            }
        }

        const loadMoreMessages = async () => {
            if (isLoadingMore || !hasMoreMessages || !selectedChat.value) return
            isLoadingMore = true

            try {
                const older = await GetChatMessages(selectedChat.value.id, {
                    limit: MESSAGES_PAGE_SIZE,
                    offset: messages.value.length,
                })

                if (older && older.length > 0) {
                    messages.value = [...older.reverse(), ...messages.value]
                    hasMoreMessages = older.length >= MESSAGES_PAGE_SIZE
                } else {
                    hasMoreMessages = false
                }
            } catch (err) {
                console.error("Ошибка загрузки старых сообщений:", err)
            } finally {
                isLoadingMore = false
            }
        }

        const closeChat = () => {
            isEmojiPickerVisible.value = false
            selectedChat.value = null
            messages.value = []
            replyToMessage.value = null
            editingMessage.value = null
            contextMenu.value.visible = false
            resetScrollDownState()
        }

        const selectChat = async (chat) => {
            selectedChat.value = chat
            await loadMessages(chat.id)
            chat.unreadCount = 0
        }

        const openChatById = async (chatId) => {
            const chat = chats.value.find(c => c.id === chatId)
            if (!chat) return

            try {
                await selectChat(chat)
            } catch (err) {
                showError(err)
            }
        }

        const sendMessage = async () => {
            if (!newMessage.value.trim() || !selectedChat.value) return

            if (editingMessage.value) {
                const msg = editingMessage.value
                const text = newMessage.value
                cancelEdit()

                try {
                    await UpdateMessage(msg.id, text)
                } catch (err) {
                    showError(err)
                    newMessage.value = text
                    editingMessage.value = msg
                }

                return
            }

            const text = newMessage.value
            const reply = replyToMessage.value
            newMessage.value = ''
            replyToMessage.value = null

            try {
                const replyId = reply ? reply.id : null
                await SendMessage(selectedChat.value.id, text, replyId)

                messages.value.forEach(m => m.isRead = true)
            } catch (err) {
                showError(err)
                newMessage.value = text
                replyToMessage.value = reply
            }
        }

        const handleNewMessage = async (message) => {
            try {
                if (!isWindowFocused && (webPushConsents.value & CONSENT_NEW_MESSAGE) !== 0) {
                    ShowNotification(message.sender.username, message.text, message.chatId)
                        .catch(err => console.error('Ошибка системного уведомления:', err))
                }

                if (selectedChat.value?.id === message.chatId) {
                    const isOwn = message.sender.id === currentUser.value?.id
                    const wasAtBottom = isAtBottom.value

                    messages.value.push({...message, isRead: false})
                    await nextTick()

                    if (isOwn || wasAtBottom) {
                        scrollToBottom()
                    } else {
                        unreadMessageIds.value.add(message.id)
                    }
                } else {
                    emit('new-message-notification', {
                        sender: message.sender.username,
                        text: message.text,
                        chatId: message.chatId,
                    })
                }

                loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))
            } catch (err) {
                console.error("Ошибка обработки нового сообщения:", err)
            }
        }

        const handleChatsUpdated = (updatedChats) => {
            chats.value = updatedChats
        }

        const scrollToBottom = () => {
            if (messagesListRef.value) {
                messagesListRef.value.scrollTop = messagesListRef.value.scrollHeight
            }
        }

        const onScrollDownClick = () => {
            if (messagesListRef.value) {
                messagesListRef.value.scrollTo({
                    top: messagesListRef.value.scrollHeight,
                    behavior: 'smooth',
                })
            }
        }

        const resetScrollDownState = () => {
            unreadMessageIds.value.clear()
            isAtBottom.value = true
            if (lastMessageObserver) {
                lastMessageObserver.disconnect()
            }
        }

        const getOtherMember = (chat) => {
            if (chat.type !== CHAT_TYPE.PRIVATE) return null
            return chat.members.find(m => m.id !== currentUser.value?.id) ?? null
        }

        const getChatTitle = (chat) => {
            if (chat.title) return chat.title

            const otherMember = getOtherMember(chat)
            if (otherMember) return otherMember.username

            return `Чат #${chat.id}`
        }

        const getChatAvatarPath = (chat) => {
            if (chat.type !== CHAT_TYPE.PRIVATE) return null
            const otherMember = getOtherMember(chat)
            return otherMember?.avatarPath || null
        }

        const openAvatarZoom = (src) => {
            if (src) avatarZoomSrc.value = src
        }

        const getLastMessagePreview = (chat) => {
            if (!chat.messages || chat.messages.length === 0) return null

            const lastMessage = chat.messages[chat.messages.length - 1]
            const senderName = lastMessage.sender.id === currentUser.value?.id
                ? 'Вы'
                : lastMessage.sender.username

            return `${senderName}: ${lastMessage.text}`
        }

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

        const openChatInfo = () => {
            if (!selectedChat.value) return
            openMemberProfile(selectedChat.value)
        }

        const returnToGroupChat = ref(null)

        const openGroupMemberProfile = (member) => {
            returnToGroupChat.value = selectedGroupChat.value
            selectedGroupChat.value = null
            selectedMember.value = member
        }

        const closeMemberProfile = () => {
            selectedMember.value = null
            if (returnToGroupChat.value) {
                selectedGroupChat.value = returnToGroupChat.value
                returnToGroupChat.value = null
            }
        }

        const getSenderName = (message) => {
            return message.sender.id === currentUser.value?.id ? 'Вы' : message.sender.username
        }

        const formatTime = (dateStr) => {
            return new Date(dateStr).toLocaleString('ru-RU')
        }

        const formatDate = (dateStr) => {
            return new Date(dateStr).toLocaleDateString('ru-RU', {
                day: 'numeric',
                month: 'long',
                year: 'numeric',
            })
        }

        const isFirstUnread = (message, index) => {
            if (message.isRead || message.sender.id === currentUser.value?.id) return false
            if (index === 0) return true

            const prev = messages.value[index - 1]
            return prev.isRead || prev.sender.id === currentUser.value?.id
        }

        const handleMessageDeleted = (payload) => {
            if (selectedChat.value?.id === payload.chatId) {
                const idx = messages.value.findIndex(m => m.id === payload.messageId)
                if (idx >= 0) {
                    messages.value.splice(idx, 1)
                    unreadMessageIds.value.delete(payload.messageId)
                } else {
                    console.warn('message_deleted: сообщение не найдено в текущем списке', payload.messageId)
                }
            }

            loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))
        }

        const handleMessageEdited = async (payload) => {
            if (selectedChat.value?.id === payload.chatId) {
                try {
                    const updated = await GetMessageByID(payload.messageId)
                    const idx = messages.value.findIndex(m => m.id === payload.messageId)
                    if (idx >= 0) {
                        messages.value[idx].text = updated.text
                        messages.value[idx].updatedAt = updated.updatedAt
                    }
                } catch (err) {
                    console.error('handleMessageEdited error:', err)
                }
            }

            loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))
        }

        const cancelReply = () => {
            replyToMessage.value = null
        }

        const cancelEdit = () => {
            editingMessage.value = null
            newMessage.value = ''
        }

        const editContextMessage = () => {
            const msg = contextMenu.value.message
            contextMenu.value.visible = false
            if (!msg) return

            replyToMessage.value = null
            editingMessage.value = msg
            newMessage.value = msg.text
            if (textareaRef.value) textareaRef.value.focus()
        }

        const scrollToMessage = async (messageId) => {
            const el = document.querySelector(`.message-bubble[data-message-id="${messageId}"]`)
            if (el) {
                el.scrollIntoView({ behavior: 'smooth', block: 'center' })
                highlightedMessageId.value = messageId
                setTimeout(() => { highlightedMessageId.value = null }, HIGHLIGHT_DURATION_MS)
            }
        }

        const openContextMenu = (event, message) => {
            const menuWidth = 200
            const menuHeight = 160
            const x = Math.min(event.clientX, window.innerWidth - menuWidth)
            const y = Math.min(event.clientY, window.innerHeight - menuHeight)

            contextMenu.value = {
                visible: true,
                x,
                y,
                message,
                deleteExpanded: false,
            }
        }

        const closeContextMenu = () => {
            contextMenu.value.visible = false
        }

        const replyToContextMessage = () => {
            replyToMessage.value = contextMenu.value.message
            contextMenu.value.visible = false
            if (textareaRef.value) textareaRef.value.focus()
        }

        const copyContextMessage = () => {
            if (contextMenu.value.message) {
                navigator.clipboard.writeText(contextMenu.value.message.text).catch(console.error)
            }
            contextMenu.value.visible = false
        }

        const deleteContextMessage = async (forAll) => {
            const messageId = contextMenu.value.message?.id
            contextMenu.value.visible = false
            if (!messageId) return

            try {
                await DeleteMessage(messageId, forAll)

                loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))
            } catch (err) {
                showError(err)
            }
        }

        const insertEmoji = async (emoji) => {
            const textarea = textareaRef.value
            if (!textarea) {
                newMessage.value += emoji
                return
            }

            const start = textarea.selectionStart
            const end = textarea.selectionEnd
            const text = newMessage.value
            newMessage.value = text.slice(0, start) + emoji + text.slice(end)

            await nextTick()
            const cursorPos = start + emoji.length
            textarea.selectionStart = cursorPos
            textarea.selectionEnd = cursorPos
            textarea.focus()
        }

        const toggleTheme = async () => {
            try {
                const newTheme = await ToggleTheme()
                isDarkTheme.value = newTheme === THEME.DARK
                const themeName = isDarkTheme.value ? 'dark' : 'light'
                document.documentElement.setAttribute('data-bs-theme', themeName)
            } catch (err) {
                showError(err)
            }
        }

        let scrollHandler = null

        onMounted(async () => {
            try {
                currentUser.value = await GetCurrentUser()
                const theme = await GetTheme()
                isDarkTheme.value = theme === THEME.DARK
                await reloadSettings()
                await loadChats()
                await StartListening()
            } catch (err) {
                console.error("Ошибка инициализации:", err)
            }

            window.runtime.EventsOn(WAILS_EVENT.NEW_MESSAGE, handleNewMessage)
            window.runtime.EventsOn(WAILS_EVENT.MESSAGE_DELETED, handleMessageDeleted)
            window.runtime.EventsOn(WAILS_EVENT.MESSAGE_EDITED, handleMessageEdited)
            window.runtime.EventsOn(WAILS_EVENT.CHATS_UPDATED, handleChatsUpdated)
            window.runtime.EventsOn(WAILS_EVENT.OPEN_CHAT, (chatId) => {
                openChatById(chatId).catch(err => console.error('Ошибка открытия чата из уведомления:', err))
            })

            window.addEventListener('focus', onWindowFocus)
            window.addEventListener('blur', onWindowBlur)
            window.addEventListener('click', closeContextMenu)
        })

        onUnmounted(() => {
            StopListening().catch(err => console.error("Ошибка остановки слушателя:", err))
            window.runtime.EventsOff(WAILS_EVENT.NEW_MESSAGE)
            window.runtime.EventsOff(WAILS_EVENT.MESSAGE_DELETED)
            window.runtime.EventsOff(WAILS_EVENT.MESSAGE_EDITED)
            window.runtime.EventsOff(WAILS_EVENT.CHATS_UPDATED)
            window.runtime.EventsOff(WAILS_EVENT.OPEN_CHAT)

            window.removeEventListener('focus', onWindowFocus)
            window.removeEventListener('blur', onWindowBlur)
            window.removeEventListener('click', closeContextMenu)
        })

        watch(messagesListRef, (el, _, onCleanup) => {
            if (!el) {
                if (lastMessageObserver) {
                    lastMessageObserver.disconnect()
                    lastMessageObserver = null
                }
                return
            }

            scrollHandler = async () => {
                if (el.scrollTop <= 10 && !isLoadingMore && hasMoreMessages && selectedChat.value) {
                    const prevHeight = el.scrollHeight
                    await loadMoreMessages()
                    await nextTick()
                    el.scrollTop = el.scrollHeight - prevHeight
                }
            }

            el.addEventListener('scroll', scrollHandler)

            lastMessageObserver = new IntersectionObserver((entries) => {
                const last = entries[entries.length - 1]
                isAtBottom.value = last.isIntersecting
                if (isAtBottom.value) {
                    unreadMessageIds.value.clear()
                }
            }, { root: el, threshold: 0.1 })

            onCleanup(() => {
                el.removeEventListener('scroll', scrollHandler)
                if (lastMessageObserver) {
                    lastMessageObserver.disconnect()
                    lastMessageObserver = null
                }
            })
        })

        watch(() => messages.value.length, async () => {
            await nextTick()
            if (!lastMessageObserver || !messagesListRef.value) return
            const bubbles = messagesListRef.value.querySelectorAll('.message-bubble')
            const last = bubbles[bubbles.length - 1]
            lastMessageObserver.disconnect()
            if (last) lastMessageObserver.observe(last)
        }, { flush: 'post' })

        return {
            chats,
            selectedChat,
            messages,
            currentUser,
            newMessage,
            messagesListRef,
            textareaRef,
            isEmojiPickerVisible,
            showEmojiPicker,
            scheduleEmojiClose,
            selectedMember,
            selectedGroupChat,
            openChatInfo,
            openGroupMemberProfile,
            closeMemberProfile,
            insertEmoji,
            closeChat,
            selectChat,
            sendMessage,
            getChatTitle,
            getChatAvatarPath,
            getLastMessagePreview,
            getOtherMember,
            openMemberProfile,
            openAvatarZoom,
            avatarZoomSrc,
            getSenderName,
            formatTime,
            formatDate,
            isFirstUnread,
            loadChats,
            openChatById,
            toggleTheme,
            isDarkTheme,
            reloadSettings,
            replyToMessage,
            editingMessage,
            highlightedMessageId,
            contextMenu,
            cancelReply,
            cancelEdit,
            editContextMessage,
            scrollToMessage,
            openContextMenu,
            replyToContextMessage,
            copyContextMessage,
            deleteContextMessage,
            isAtBottom,
            unreadMessageIds,
            onScrollDownClick,
        }
    }
}
