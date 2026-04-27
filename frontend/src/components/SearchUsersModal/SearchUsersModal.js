import {ref} from 'vue'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'

export default {
    name: 'SearchUsersModal', emits: ['close'],

    setup() {
        const searchQuery = ref('')
        const searchResults = ref([])
        const searched = ref(false)

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

            try {
                const users = await SearchUsers({
                    username: searchQuery.value,
                }, null)
                searchResults.value = users
            } catch (err) {
                searched.value = false

                alert(`Ошибка поиска: ${err}`)
            }
        }

        return {
            searchQuery, searchResults, searched, debouncedSearch
        }
    }
}