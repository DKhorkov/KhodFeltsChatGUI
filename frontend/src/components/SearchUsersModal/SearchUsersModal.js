import {inject, ref} from 'vue'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'
import {SEARCH_DEBOUNCE_MS} from '../../constants'

export default {
    name: 'SearchUsersModal', emits: ['close'],

    setup() {
        const showError = inject('showError')
        const searchQuery = ref('')
        const searchResults = ref([])
        const searched = ref(false)

        let debounceTimer = null

        const debouncedSearch = () => {
            clearTimeout(debounceTimer)
            debounceTimer = setTimeout(async () => {
                if (searchQuery.value) {
                    await searchUsers()
                }
            }, SEARCH_DEBOUNCE_MS)
        }

        const searchUsers = async () => {
            searched.value = true

            try {
                searchResults.value = await SearchUsers({username: searchQuery.value}, null)
            } catch (err) {
                searched.value = false
                showError(err)
            }
        }

        return {
            searchQuery,
            searchResults,
            searched,
            debouncedSearch,
        }
    }
}
