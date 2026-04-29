import {inject, ref} from 'vue'
import {CreateChat} from '../../../wailsjs/go/create_chat/Handler'
import {SearchUsers} from '../../../wailsjs/go/search_users/Handler'
import {CHAT_TYPE, SEARCH_DEBOUNCE_MS} from '../../constants'
import {debounce} from '../../utils/debounce'

export default {
    name: 'CreateChatModal',
    emits: ['close', 'chat-created'],

    setup(props, {emit}) {
        const showError = inject('showError')
        const showInfo = inject('showInfo')
        const chatType = ref(CHAT_TYPE.PRIVATE)
        const chatTitle = ref('')
        const chatDescription = ref('')
        const searchQuery = ref('')
        const searchResults = ref([])
        const selectedUserIds = ref([])

        const searchUsers = async () => {
            if (!searchQuery.value) return

            try {
                searchResults.value = await SearchUsers({username: searchQuery.value}, null)
            } catch (err) {
                showError(err)
            }
        }

        const handleSearchInput = debounce(searchUsers, SEARCH_DEBOUNCE_MS)

        const createChat = async () => {
            if (selectedUserIds.value.length === 0) {
                showInfo('Укажите хотя бы одного участника')
                return
            }

            try {
                await CreateChat({
                    type: chatType.value,
                    memberIDs: selectedUserIds.value,
                    title: chatType.value === CHAT_TYPE.GROUP ? chatTitle.value : null,
                    description: chatType.value === CHAT_TYPE.GROUP ? chatDescription.value : null,
                })

                emit('chat-created')
            } catch (err) {
                showError(err)
            }
        }

        return {
            CHAT_TYPE,
            chatType,
            chatTitle,
            chatDescription,
            searchQuery,
            searchResults,
            selectedUserIds,
            handleSearchInput,
            createChat,
        }
    }
}
