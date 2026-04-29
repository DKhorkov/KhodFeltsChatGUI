import {inject, nextTick, onMounted, onUnmounted, ref, watch} from 'vue'
import {
    GetChatMessages,
    GetCurrentUser,
    GetUserChats,
    SendMessage,
    StartListening,
    StopListening
} from '../../../wailsjs/go/chat/Handler'
import {GetTheme, ToggleTheme} from '../../../wailsjs/go/settings/Handler'
import {CHAT_TYPE, MESSAGES_PAGE_SIZE, THEME, WAILS_EVENT} from '../../constants'
import EmojiPicker from '../EmojiPicker/EmojiPicker.vue'

export default {
    name: 'ChatView',
    components: {EmojiPicker},
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

        let isLoadingMore = false
        let hasMoreMessages = true

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

        const selectChat = async (chat) => {
            selectedChat.value = chat
            await loadMessages(chat.id)
            chat.isRead = true
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

            const text = newMessage.value
            newMessage.value = ''

            try {
                await SendMessage(selectedChat.value.id, text)

                messages.value.forEach(m => m.isRead = true)

                messages.value.push({
                    id: Date.now(),
                    text,
                    chatId: selectedChat.value.id,
                    createdAt: new Date().toISOString(),
                    sender: {
                        id: currentUser.value?.id,
                        username: currentUser.value?.username,
                    },
                })

                await nextTick()
                scrollToBottom()
            } catch (err) {
                showError(err)
                newMessage.value = text
            }
        }

        const handleNewMessage = async (message) => {
            try {
                if (selectedChat.value?.id === message.chatId) {
                    messages.value.push({...message, isRead: false})
                    await nextTick()
                    scrollToBottom()
                    return
                }

                loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))

                emit('new-message-notification', {
                    text: `${message.sender.username}: ${message.text}`,
                    chatId: message.chatId,
                })
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

        const getChatTitle = (chat) => {
            if (chat.title) return chat.title

            if (chat.type === CHAT_TYPE.PRIVATE) {
                const otherMember = chat.members.find(m => m.id !== currentUser.value?.id)
                if (otherMember) return otherMember.username
            }

            return `Чат #${chat.id}`
        }

        const getSenderName = (message) => {
            return message.sender.id === currentUser.value?.id ? 'Вы' : message.sender.username
        }

        const formatTime = (dateStr) => {
            return new Date(dateStr).toLocaleString('ru-RU')
        }

        const isFirstUnread = (message, index) => {
            if (message.isRead || message.sender.id === currentUser.value?.id) return false
            if (index === 0) return true

            const prev = messages.value[index - 1]
            return prev.isRead || prev.sender.id === currentUser.value?.id
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
            if (Notification.permission !== 'granted') {
                await Notification.requestPermission()
            }

            try {
                currentUser.value = await GetCurrentUser()
                const theme = await GetTheme()
                isDarkTheme.value = theme === THEME.DARK
                await loadChats()
                await StartListening()
            } catch (err) {
                console.error("Ошибка инициализации:", err)
            }

            window.runtime.EventsOn(WAILS_EVENT.NEW_MESSAGE, handleNewMessage)
            window.runtime.EventsOn(WAILS_EVENT.CHATS_UPDATED, handleChatsUpdated)
        })

        onUnmounted(() => {
            StopListening().catch(err => console.error("Ошибка остановки слушателя:", err))
            window.runtime.EventsOff(WAILS_EVENT.NEW_MESSAGE)
            window.runtime.EventsOff(WAILS_EVENT.CHATS_UPDATED)
        })

        watch(messagesListRef, (el, _, onCleanup) => {
            if (!el) return

            scrollHandler = async () => {
                if (el.scrollTop <= 10 && !isLoadingMore && hasMoreMessages && selectedChat.value) {
                    const prevHeight = el.scrollHeight
                    await loadMoreMessages()
                    await nextTick()
                    el.scrollTop = el.scrollHeight - prevHeight
                }
            }

            el.addEventListener('scroll', scrollHandler)
            onCleanup(() => el.removeEventListener('scroll', scrollHandler))
        })

        return {
            chats,
            selectedChat,
            messages,
            currentUser,
            newMessage,
            messagesListRef,
            textareaRef,
            isEmojiPickerVisible,
            insertEmoji,
            selectChat,
            sendMessage,
            getChatTitle,
            getSenderName,
            formatTime,
            isFirstUnread,
            loadChats,
            openChatById,
            toggleTheme,
            isDarkTheme,
        }
    }
}
