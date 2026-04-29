import {inject, ref} from 'vue'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'
import {SEARCH_DEBOUNCE_MS} from '../../constants'
import {debounce} from '../../utils/debounce'

export default {
    name: 'SearchUsersModal',
    emits: ['close'],

    setup() {
        const showError = inject('showError')
        const searchQuery = ref('')
        const searchResults = ref([])
        const hasSearched = ref(false)
        const selectedUser = ref(null)

        const searchUsers = async () => {
            if (!searchQuery.value) return

            hasSearched.value = true

            try {
                searchResults.value = await SearchUsers({username: searchQuery.value}, null)
            } catch (err) {
                hasSearched.value = false
                showError(err)
            }
        }

        const handleSearchInput = debounce(searchUsers, SEARCH_DEBOUNCE_MS)

        const formatDate = (dateStr) => {
            return new Date(dateStr).toLocaleDateString('ru-RU', {
                day: 'numeric',
                month: 'long',
                year: 'numeric',
            })
        }

        return {
            searchQuery,
            searchResults,
            hasSearched,
            selectedUser,
            handleSearchInput,
            formatDate,
        }
    }
}
