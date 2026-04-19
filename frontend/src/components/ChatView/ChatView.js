import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'

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
      chats.value = await window.go.main.ChatHandler.GetUserChats(0, 0)
    }

    const loadMessages = async (chatId) => {
      const msgs = await window.go.main.ChatHandler.GetChatMessages(chatId, 10, 0)
      messages.value = msgs.reverse()
      hasMoreMessages = msgs.length >= 10

      await nextTick()
      scrollToBottom()
    }

    const loadMoreMessages = async () => {
      if (loadMoreLock || !hasMoreMessages || !currentChat.value) return

      loadMoreLock = true
      const offset = messages.value.length

      const olderMessages = await window.go.main.ChatHandler.GetChatMessages(
          currentChat.value.ID,
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
      await loadMessages(chat.ID)

      // Отмечаем чат как прочитанный
      chat.IsRead = true
    }

    const sendMessage = async () => {
      if (!newMessage.value.trim() || !currentChat.value) return

      await window.go.main.ChatHandler.SendMessage({
        chatId: currentChat.value.ID,
        message: newMessage.value
      })

      newMessage.value = ''
    }

    const handleNewMessage = (message) => {
      if (currentChat.value?.ID === message.ChatID) {
        messages.value.push(message)
        scrollToBottom()
      } else {
        // Обновляем список чатов для показа индикатора
        loadChats()

        // Показываем уведомление
        if (Notification.permission === 'granted') {
          new Notification('Новое сообщение', {
            body: `${message.Sender.Username}: ${message.Text}`
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
      if (chat.Title && chat.Title !== '') {
        return chat.Title
      }

      if (chat.Type === 'private') {
        const otherMember = chat.Members.find(m => m.ID !== currentUser.value?.ID)
        if (otherMember) return otherMember.Username
      }

      return `Чат #${chat.ID}`
    }

    const getSenderName = (message) => {
      if (message.Sender.ID === currentUser.value?.ID) {
        return 'Вы'
      }
      return message.Sender.Username
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
      currentUser.value = await window.go.main.ChatHandler.GetCurrentUser()

      // Загружаем чаты
      await loadChats()

      // Запускаем прослушивание сообщений
      await window.go.main.ChatHandler.StartListening()

      // Подписываемся на события
      window.runtime.EventsOn('new_message', handleNewMessage)
      window.runtime.EventsOn('chats_updated', handleChatsUpdated)
    })

    onUnmounted(() => {
      window.go.main.ChatHandler.StopListening()
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