import {ref} from 'vue'
import {CreateChat} from '../../../wailsjs/go/create_chat/Handler'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'
import {CHAT_TYPE, SEARCH_DEBOUNCE_MS} from '../../constants'

export default {
    name: 'CreateChatModal', emits: ['close', 'chat-created'],

    setup(props, {emit}) {
        const chatType = ref(CHAT_TYPE.PRIVATE)
        const chatTitle = ref('')
        const chatDescription = ref('')
        const searchQuery = ref('')
        const searchResults = ref([])
        const selectedUsers = ref([])

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
            try {
                searchResults.value = await SearchUsers({username: searchQuery.value}, null)
            } catch (err) {
                alert(err)
            }
        }

        const createChat = async () => {
            if (chatType.value === CHAT_TYPE.PRIVATE && selectedUsers.value.length === 0) {
                alert('Укажите хотя бы одного участника')
                return
            }

            try {
                await CreateChat({
                    type: chatType.value,
                    memberIDs: selectedUsers.value,
                    title: chatType.value === CHAT_TYPE.GROUP ? chatTitle.value : null,
                    description: chatType.value === CHAT_TYPE.GROUP ? chatDescription.value : null,
                })

                emit('chat-created')
            } catch (err) {
                alert(err)
            }
        }

        return {
            CHAT_TYPE,
            chatType,
            chatTitle,
            searchQuery,
            searchResults,
            selectedUsers,
            debouncedSearch,
            createChat,
        }
    }
}
