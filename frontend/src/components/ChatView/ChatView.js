import {nextTick, onMounted, onUnmounted, ref, watch} from 'vue'
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

export default {
    name: 'ChatView', emits: ['logout', 'show-create-chat', 'show-search-users', 'new-message-notification'],

    setup(props, {emit}) {
        const chats = ref([])
        const currentChat = ref(null)
        const messages = ref([])
        const currentUser = ref(null)
        const newMessage = ref('')
        const messagesList = ref(null)
        const isDarkTheme = ref(false)

        let loadMoreLock = false
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
                const msgs = await GetChatMessages(chatId, {
                    limit: MESSAGES_PAGE_SIZE,
                    offset: 0,
                })
                messages.value = msgs.reverse()
                hasMoreMessages = msgs.length >= MESSAGES_PAGE_SIZE

                await nextTick()
                scrollToBottom()
            } catch (err) {
                console.error("Ошибка загрузки сообщений:", err)
            }
        }

        const loadMoreMessages = async () => {
            if (loadMoreLock || !hasMoreMessages || !currentChat.value) return
            loadMoreLock = true

            try {
                const olderMessages = await GetChatMessages(currentChat.value.id, {
                    limit: MESSAGES_PAGE_SIZE,
                    offset: messages.value.length,
                })

                if (olderMessages && olderMessages.length > 0) {
                    messages.value = [...olderMessages.reverse(), ...messages.value]
                    hasMoreMessages = olderMessages.length >= MESSAGES_PAGE_SIZE
                } else {
                    hasMoreMessages = false
                }
            } catch (err) {
                console.error("Ошибка загрузки старых сообщений:", err)
            } finally {
                loadMoreLock = false
            }
        }

        const selectChat = async (chat) => {
            currentChat.value = chat
            await loadMessages(chat.id)
            chat.isRead = true
        }

        const openChatById = (chatId) => {
            const chat = chats.value.find(c => c.id === chatId)
            if (chat) {
                selectChat(chat)
            }
        }

        const sendMessage = async () => {
            if (!newMessage.value.trim() || !currentChat.value) return

            const text = newMessage.value
            newMessage.value = ''

            try {
                await SendMessage(currentChat.value.id, text)

                messages.value.forEach(m => m.isRead = true)

                messages.value.push({
                    id: Date.now(),
                    text,
                    chatId: currentChat.value.id,
                    createdAt: new Date().toISOString(),
                    sender: {
                        id: currentUser.value?.id,
                        username: currentUser.value?.username,
                    },
                })

                await nextTick()
                scrollToBottom()
            } catch (err) {
                alert(err)
                newMessage.value = text
            }
        }

        const handleNewMessage = (message) => {
            if (currentChat.value?.id === message.chatId) {
                messages.value.push(message)
                nextTick(() => scrollToBottom())
                return
            }

            loadChats().catch(err => console.error("Фоновое обновление чатов не удалось:", err))

            emit('new-message-notification', {
                text: `${message.sender.username}: ${message.text}`,
                chatId: message.chatId,
            })
        }

        const handleChatsUpdated = (updatedChats) => {
            chats.value = updatedChats
        }

        const scrollToBottom = () => {
            if (messagesList.value) {
                messagesList.value.scrollTop = messagesList.value.scrollHeight
            }
        }

        const getChatTitle = (chat) => {
            if (chat.title) {
                return chat.title
            }

            if (chat.type === CHAT_TYPE.PRIVATE) {
                const otherMember = chat.members.find(m => m.id !== currentUser.value?.id)
                if (otherMember) return otherMember.username
            }

            return `Чат #${chat.id}`
        }

        const getSenderName = (message) => {
            return `${message.sender.id === currentUser.value?.id ? 'Вы' : message.sender.username} `
        }

        const formatTime = (dateStr) => {
            return new Date(dateStr).toLocaleString('ru-RU')
        }

        const isUnreadDivider = (message, index) => {
            if (message.isRead || message.sender.id === currentUser.value?.id) return false
            for (let i = 0; i < index; i++) {
                const prev = messages.value[i]
                if (!prev.isRead && prev.sender.id !== currentUser.value?.id) return false
            }
            return true
        }

        const handleLogout = () => {
            emit('logout')
        }

        const toggleTheme = async () => {
            try {
                const newTheme = await ToggleTheme()
                isDarkTheme.value = newTheme === THEME.DARK
                const themeName = isDarkTheme.value ? 'dark' : 'light'
                document.documentElement.setAttribute('data-bs-theme', themeName)
            } catch (err) {
                alert(err)
            }
        }

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
            StopListening()
            window.runtime.EventsOff(WAILS_EVENT.NEW_MESSAGE)
            window.runtime.EventsOff(WAILS_EVENT.CHATS_UPDATED)
        })

        watch(messagesList, (el) => {
            if (!el) return

            const handleScroll = async () => {
                if (el.scrollTop <= 10 && !loadMoreLock && hasMoreMessages && currentChat.value) {
                    const prevHeight = el.scrollHeight
                    await loadMoreMessages()
                    await nextTick()
                    el.scrollTop = el.scrollHeight - prevHeight
                }
            }

            el.addEventListener('scroll', handleScroll)
            return () => el.removeEventListener('scroll', handleScroll)
        })

        return {
            chats,
            currentChat,
            messages,
            currentUser,
            newMessage,
            messagesList,
            selectChat,
            sendMessage,
            getChatTitle,
            getSenderName,
            formatTime,
            isUnreadDivider,
            handleLogout,
            loadChats,
            openChatById,
            toggleTheme,
            isDarkTheme,
        }
    }
}