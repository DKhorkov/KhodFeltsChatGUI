import { ref } from 'vue'
import { SearchUsers, CreateChat } from '../../../wailsjs/go/create_chat/Handler'

export default {
  name: 'CreateChatModal',
  emits: ['close', 'chat-created'],

  setup(props, { emit }) {
    const chatType = ref('private')
    const chatTitle = ref('')
    const searchQuery = ref('')
    const searchResults = ref([])
    const selectedUsers = ref([])
    const loading = ref(false)

    let debounceTimer = null

    const debouncedSearch = () => {
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        if (searchQuery.value) {
          searchUsers()
        }
      }, 500)
    }

    const searchUsers = async () => {
      try {
        const users = await SearchUsers(
            searchQuery.value,
            0,
            0
        )
        searchResults.value = users
      } catch (err) {
        console.error('Ошибка поиска:', err)
      }
    }

    const createChat = async () => {
      if (chatType.value === 'private' && selectedUsers.value.length === 0) {
        alert('Укажите хотя бы одного участника')
        return
      }

      loading.value = true

      try {
        await CreateChat({
          type: chatType.value,
          members: selectedUsers.value,
          title: chatType.value === 'group' ? chatTitle.value : null
        })

        emit('chat-created')
      } catch (err) {
        console.error('Ошибка создания чата:', err)
        alert('Ошибка создания чата')
      } finally {
        loading.value = false
      }
    }

    return {
      chatType,
      chatTitle,
      searchQuery,
      searchResults,
      selectedUsers,
      loading,
      debouncedSearch,
      createChat
    }
  }
}