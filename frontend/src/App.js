import {onMounted, ref} from 'vue'
import LoginView from './components/LoginView/LoginView.vue'
import ChatView from './components/ChatView/ChatView.vue'
import CreateChatModal from './components/CreateChatModal/CreateChatModal.vue'
import SearchUsersModal from './components/SearchUsersModal/SearchUsersModal.vue'
import ForgetPasswordModal from './components/ForgetPasswordModal/ForgetPasswordModal.vue'
import NotificationToast from './components/NotificationToast/NotificationToast.vue'

import {Authenticate, Logout} from '../wailsjs/go/auth/Handler'


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
        const currentView = ref('loading')
        const showCreateChatModal = ref(false)
        const showSearchUsersModal = ref(false)
        const showForgetPasswordModal = ref(false)
        const notification = ref(null)

        const checkSession = async () => {
            try {
                // Мы просто ждем выполнения. Если ошибки нет — значит Authenticate вернул nil
                await Authenticate()

                // Если код дошел до этой строки, значит ошибки не было
                currentView.value = 'chat'
            } catch (err) {
                console.error("Ошибка проверки сессии:", err)
                currentView.value = 'login'
            }
        }

        onMounted(() => {
            checkSession()
        })

        const handleLoginSuccess = () => {
            currentView.value = 'chat'
        }

        const handleLogout = async () => { // Оставили только один уровень
            try {
                await Logout()
            } catch (err) {
                console.error("Ошибка logout:", err)
            }

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