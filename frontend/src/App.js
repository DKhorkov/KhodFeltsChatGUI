import {ref} from 'vue'
import LoginView from './components/LoginView/LoginView.vue'
import ChatView from './components/ChatView/ChatView.vue'
import CreateChatModal from './components/CreateChatModal/CreateChatModal.vue'
import SearchUsersModal from './components/SearchUsersModal/SearchUsersModal.vue'
import ForgetPasswordModal from './components/ForgetPasswordModal/ForgetPasswordModal.vue'
import NotificationToast from './components/NotificationToast/NotificationToast.vue'

export default {
    name: 'App',
    components: {
        LoginView,
        ChatView,
        CreateChatModal,
        SearchUsersModal,
        ForgetPasswordModal,
        NotificationToast
    },

    setup() {
        const currentView = ref('login')
        const showCreateChatModal = ref(false)
        const showSearchUsersModal = ref(false)
        const showForgetPasswordModal = ref(false)
        const notification = ref(null)

        const handleLoginSuccess = () => {
            currentView.value = 'chat'
        }

        const handleLogout = () => {
            currentView.value = 'login'
        }

        const handleChatCreated = () => {
            showCreateChatModal.value = false
            notification.value = 'Чат успешно создан!'
            setTimeout(() => {
                notification.value = null
            }, 3000)
        }

        return {
            currentView,
            showCreateChatModal,
            showSearchUsersModal,
            showForgetPasswordModal,
            notification,
            handleLoginSuccess,
            handleLogout,
            handleChatCreated
        }
    }
}