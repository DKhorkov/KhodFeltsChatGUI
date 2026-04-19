import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { GetUserChats, GetChatMessages, SendMessage, GetCurrentUser, StartListening, StopListening } from '../../../wailsjs/go/chat/Handler'

export default {
  name: 'ChatView',
  emits: ['logout', 'show-create-chat', 'show-search-users'],

  setup(props, { emit }) {
    const chats = ref([])
    const currentChat = ref(null)
    const messages = ref([])
    const currentUser = ref(null)
    const newMessage = ref('')
    const messagesList = ref(null)

    let loadMoreLock = false
    let hasMoreMessages = true

    const loadChats = async () => {
      chats.value = await GetUserChats(0, 0)
    }

    const loadMessages = async (chatId) => {
      const msgs = await GetChatMessages(chatId, 10, 0)
      messages.value = msgs.reverse()
      hasMoreMessages = msgs.length >= 10

      await nextTick()
      scrollToBottom()
    }

    const loadMoreMessages = async () => {
      if (loadMoreLock || !hasMoreMessages || !currentChat.value) return

      loadMoreLock = true
      const offset = messages.value.length

      const olderMessages = await GetChatMessages(
          currentChat.value.id,
          10,
          offset
      )

      if (olderMessages.length > 0) {
        messages.value = [...olderMessages.reverse(), ...messages.value]
        hasMoreMessages = olderMessages.length >= 10
      } else {
        hasMoreMessages = false
      }

      loadMoreLock = false
    }

    const selectChat = async (chat) => {
      currentChat.value = chat
      await loadMessages(chat.id)

      // Отмечаем чат как прочитанный
      chat.IsRead = true
    }

    const sendMessage = async () => {
      if (!newMessage.value.trim() || !currentChat.value) return

      await SendMessage({
        chatId: currentChat.value.id,
        message: newMessage.value
      })

      newMessage.value = ''
    }

    const handleNewMessage = (message) => {
      if (currentChat.value?.id === message.chatId) {
        messages.value.push(message)
        scrollToBottom()
      } else {
        // Обновляем список чатов для показа индикатора
        loadChats()

        // Показываем уведомление
        if (Notification.permission === 'granted') {
          new Notification('Новое сообщение', {
            body: `${message.sender.username}: ${message.Text}`
          })
        }
      }
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
      if (chat.title && chat.title !== '') {
        return chat.title
      }

      if (chat.type === 'private') {
        const otherMember = chat.members.find(m => m.id !== currentUser.value?.id)
        if (otherMember) return otherMember.username
      }

      return `Чат #${chat.id}`
    }

    const getSenderName = (message) => {
      if (message.sender.id === currentUser.value?.id) {
        return 'Вы'
      }
      return message.sender.username
    }

    const formatTime = (dateStr) => {
      const date = new Date(dateStr)
      return date.toLocaleString('ru-RU')
    }

    const handleLogout = () => {
      emit('logout')
    }

    onMounted(async () => {
      // Запрашиваем разрешение на уведомления
      if (Notification.permission !== 'granted') {
        await Notification.requestPermission()
      }

      // Загружаем текущего пользователя
      currentUser.value = await GetCurrentUser()

      // Загружаем чаты
      await loadChats()

      // Запускаем прослушивание сообщений
      await StartListening()

      // Подписываемся на события
      window.runtime.EventsOn('new_message', handleNewMessage)
      window.runtime.EventsOn('chats_updated', handleChatsUpdated)
    })

    onUnmounted(() => {
      StopListening()
      window.runtime.EventsOff('new_message')
      window.runtime.EventsOff('chats_updated')
    })

    // Наблюдатель за скроллом для подгрузки сообщений
    watch(messagesList, (el) => {
      if (!el) return

      const handleScroll = () => {
        if (el.scrollTop === 0) {
          loadMoreMessages()
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
      handleLogout
    }
  }
}