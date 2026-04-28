import {onMounted, provide, ref} from 'vue'
import LoginView from './components/LoginView/LoginView.vue'
import ChatView from './components/ChatView/ChatView.vue'
import CreateChatModal from './components/CreateChatModal/CreateChatModal.vue'
import SearchUsersModal from './components/SearchUsersModal/SearchUsersModal.vue'
import ForgetPasswordModal from './components/ForgetPasswordModal/ForgetPasswordModal.vue'
import NotificationToast from './components/NotificationToast/NotificationToast.vue'
import AlertModal from './components/AlertModal/AlertModal.vue'

import {Authenticate, Logout} from '../wailsjs/go/auth/Handler'
import {GetTheme} from '../wailsjs/go/settings/Handler'
import {NOTIFICATION_DURATION_MS, THEME, VIEW} from './constants'

export default {
    name: 'App', components: {
        LoginView, ChatView, CreateChatModal, SearchUsersModal, ForgetPasswordModal, NotificationToast, AlertModal
    },

    setup() {
        const currentView = ref(VIEW.LOADING)
        const chatViewComponent = ref(null)
        const showCreateChatModal = ref(false)
        const showSearchUsersModal = ref(false)
        const showForgetPasswordModal = ref(false)
        const notifications = ref([])
        let notificationId = 0
        const forgetPasswordMessage = ref('')
        const alertModal = ref(null)

        const showError = (message) => {
            alertModal.value = { message: String(message), type: 'error' }
        }

        const showInfo = (message) => {
            alertModal.value = { message: String(message), type: 'info' }
        }

        const closeAlert = () => {
            alertModal.value = null
        }

        // Механизм — через provide/inject:
        // - App.js предоставляет showError() и showInfo()
        // - Любой дочерний компонент получает их через inject() — не нужно пробрасывать emit через всю цепочку
        provide('showError', showError)
        provide('showInfo', showInfo)

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

        // Используется один раз в жизненном цикле, когда компонент впервые отрисовывается
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

        const addNotification = (message, chatId = null) => {
            const id = ++notificationId
            notifications.value.push({ id, message, chatId })

            setTimeout(() => {
                removeNotification(id)
            }, NOTIFICATION_DURATION_MS)
        }

        const removeNotification = (id) => {
            notifications.value = notifications.value.filter(n => n.id !== id)
        }

        const handleChatCreated = () => {
            showCreateChatModal.value = false
            addNotification('Чат успешно создан!')

            if (chatViewComponent.value) {
                chatViewComponent.value.loadChats().catch(err => console.error("Ошибка обновления чатов:", err))
            }
        }

        const handleNewMessageNotification = ({ text, chatId }) => {
            addNotification(text, chatId)
        }

        const handleNotificationClick = (id) => {
            const n = notifications.value.find(n => n.id === id)
            if (n?.chatId && chatViewComponent.value) {
                chatViewComponent.value.openChatById(n.chatId)
            }
            removeNotification(id)
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
            notifications,
            removeNotification,
            handleLoginSuccess,
            handleLogout,
            handleChatCreated,
            forgetPasswordMessage,
            handleShowForgetPassword,
            handleNewMessageNotification,
            handleNotificationClick,
            alertModal,
            closeAlert,
        }
    }
}