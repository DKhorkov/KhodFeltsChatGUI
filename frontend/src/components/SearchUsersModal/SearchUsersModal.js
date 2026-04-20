import {ref} from 'vue'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'

export default {
    name: 'SearchUsersModal', emits: ['close'],

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
                const users = await SearchUsers({
                    username: searchQuery.value,
                }, null)
                searchResults.value = users
            } catch (err) {
                console.error('Ошибка поиска:', err)
            } finally {
                loading.value = false
            }
        }

        return {
            searchQuery, searchResults, searched, loading, debouncedSearch
        }
    }
}