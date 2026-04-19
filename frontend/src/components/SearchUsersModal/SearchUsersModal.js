import { ref } from 'vue'

export default {
  name: 'SearchUsersModal',
  emits: ['close'],

  setup() {
    const searchQuery = ref('')
    const searchResults = ref([])
    const searched = ref(false)
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
      searched.value = true
      loading.value = true

      try {
        const users = await window.go.main.SearchUsersHandler.SearchUsers(
            searchQuery.value,
            0,
            0
        )
        searchResults.value = users
      } catch (err) {
        console.error('Ошибка поиска:', err)
      } finally {
        loading.value = false
      }
    }

    return {
      searchQuery,
      searchResults,
      searched,
      loading,
      debouncedSearch
    }
  }
}