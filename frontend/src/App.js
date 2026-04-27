import {onMounted, ref} from 'vue'
import LoginView from './components/LoginView/LoginView.vue'
import ChatView from './components/ChatView/ChatView.vue'
import CreateChatModal from './components/CreateChatModal/CreateChatModal.vue'
import SearchUsersModal from './components/SearchUsersModal/SearchUsersModal.vue'
import ForgetPasswordModal from './components/ForgetPasswordModal/ForgetPasswordModal.vue'
import NotificationToast from './components/NotificationToast/NotificationToast.vue'

import {Authenticate, Logout} from '../wailsjs/go/auth/Handler'
import {GetTheme} from '../wailsjs/go/settings/Handler'
import {NOTIFICATION_DURATION_MS, THEME, VIEW} from './constants'

export default {
    name: 'App', components: {
        LoginView, ChatView, CreateChatModal, SearchUsersModal, ForgetPasswordModal, NotificationToast
    },

    setup() {
        const currentView = ref(VIEW.LOADING)
        const chatViewComponent = ref(null)
        const showCreateChatModal = ref(false)
        const showSearchUsersModal = ref(false)
        const showForgetPasswordModal = ref(false)
        const notification = ref(null)
        const forgetPasswordMessage = ref('')

        const checkSession = async () => {
            try {
                await Authenticate()
                currentView.value = VIEW.CHAT
            } catch (err) {
                console.error("Ошибка проверки сессии:", err)
                currentView.value = VIEW.LOGIN
            }
        }

        const applyTheme = (themeType) => {
            const themeName = themeType === THEME.DARK ? 'dark' : 'light'
            document.documentElement.setAttribute('data-bs-theme', themeName)
        }

        onMounted(async () => {
            await checkSession()

            try {
                const theme = await GetTheme()
                applyTheme(theme)
            } catch (err) {
                console.error("Ошибка загрузки темы:", err)
            }
        })

        const handleLoginSuccess = () => {
            currentView.value = VIEW.CHAT
        }

        const handleLogout = async () => {
            try {
                await Logout()
            } catch (err) {
                console.error("Ошибка logout:", err)
            }

            currentView.value = VIEW.LOGIN
        }

        const handleChatCreated = () => {
            showCreateChatModal.value = false
            notification.value = 'Чат успешно создан!'

            if (chatViewComponent.value) {
                chatViewComponent.value.loadChats().catch(err => console.error("Ошибка обновления чатов:", err))
            }

            setTimeout(() => {
                notification.value = null
            }, NOTIFICATION_DURATION_MS)
        }

        const handleShowForgetPassword = (msg) => {
            forgetPasswordMessage.value = msg || 'Инструкции отправлены на почту'
            showForgetPasswordModal.value = true
        }

        return {
            VIEW,
            currentView,
            chatViewComponent,
            showCreateChatModal,
            showSearchUsersModal,
            showForgetPasswordModal,
            notification,
            handleLoginSuccess,
            handleLogout,
            handleChatCreated,
            forgetPasswordMessage,
            handleShowForgetPassword,
        }
    }
}