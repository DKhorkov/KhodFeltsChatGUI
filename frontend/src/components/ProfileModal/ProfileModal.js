import {computed, inject, onMounted, onUnmounted, ref} from 'vue'
import {ChangePassword} from '../../../wailsjs/go/auth/Handler'
import {UpdateUser, UpdateAvatar, DeleteAvatar} from '../../../wailsjs/go/users/Handler'
import {GetSettings, UpdateSettings} from '../../../wailsjs/go/settings/Handler'

const CONSENT_NEW_MESSAGE = 1

export default {
    name: 'ProfileModal',
    props: {
        user: {
            type: Object,
            required: true,
        },
        isDarkTheme: {
            type: Boolean,
            required: true,
        },
    },
    emits: ['close', 'toggle-theme', 'logout', 'user-updated'],

    setup(props, {emit}) {
        const showError = inject('showError')
        const showInfo = inject('showInfo')

        const isEditProfileOpen = ref(false)
        const editUsername = ref('')

        const isChangePasswordOpen = ref(false)
        const oldPassword = ref('')
        const newPassword = ref('')
        const confirmPassword = ref('')

        const isAvatarMenuOpen = ref(false)
        const avatarFileInput = ref(null)

        const isNotificationsOpen = ref(false)
        const emailConsents = ref(0)
        const webPushConsents = ref(0)

        const emailNewMessageConsent = computed(() => (emailConsents.value & CONSENT_NEW_MESSAGE) !== 0)
        const webPushNewMessageConsent = computed(() => (webPushConsents.value & CONSENT_NEW_MESSAGE) !== 0)

        const loadSettings = async () => {
            try {
                const settings = await GetSettings()
                emailConsents.value = settings.emailConsents
                webPushConsents.value = settings.webPushConsents
            } catch (err) {
                console.error('Ошибка загрузки настроек:', err)
            }
        }

        const toggleEmailConsent = async () => {
            try {
                const newValue = emailConsents.value ^ CONSENT_NEW_MESSAGE
                const settings = await UpdateSettings({
                    theme: props.isDarkTheme ? 1 : 0,
                    emailConsents: newValue,
                    webPushConsents: webPushConsents.value,
                })
                emailConsents.value = settings.emailConsents
                webPushConsents.value = settings.webPushConsents
            } catch (err) {
                showError(err)
            }
        }

        const toggleWebPushConsent = async () => {
            try {
                const newValue = webPushConsents.value ^ CONSENT_NEW_MESSAGE
                const settings = await UpdateSettings({
                    theme: props.isDarkTheme ? 1 : 0,
                    emailConsents: emailConsents.value,
                    webPushConsents: newValue,
                })
                emailConsents.value = settings.emailConsents
                webPushConsents.value = settings.webPushConsents
            } catch (err) {
                showError(err)
            }
        }

        const formatDate = (dateStr) => {
            return new Date(dateStr).toLocaleDateString('ru-RU', {
                day: 'numeric',
                month: 'long',
                year: 'numeric',
            })
        }

        const changePassword = async () => {
            if (!oldPassword.value || !newPassword.value || !confirmPassword.value) {
                showInfo('Заполните все поля')
                return
            }

            if (newPassword.value !== confirmPassword.value) {
                showError('Пароли не совпадают')
                return
            }

            try {
                await ChangePassword({
                    oldPassword: oldPassword.value,
                    newPassword: newPassword.value,
                })
                showInfo('Пароль успешно изменён')
                oldPassword.value = ''
                newPassword.value = ''
                confirmPassword.value = ''
                isChangePasswordOpen.value = false
            } catch (err) {
                showError(err)
            }
        }

        const updateUser = async () => {
            try {
                const dto = {}
                if (editUsername.value) {
                    dto.username = editUsername.value
                }

                const updatedUser = await UpdateUser(dto)
                showInfo('Профиль успешно обновлён')
                emit('user-updated', updatedUser)
                editUsername.value = ''
                isEditProfileOpen.value = false
            } catch (err) {
                showError(err)
            }
        }

        const toggleAvatarMenu = () => {
            isAvatarMenuOpen.value = !isAvatarMenuOpen.value
        }

        const closeAvatarMenu = () => {
            isAvatarMenuOpen.value = false
        }

        const triggerFileInput = () => {
            isAvatarMenuOpen.value = false
            if (avatarFileInput.value) avatarFileInput.value.click()
        }

        const uploadAvatar = async () => {
            const file = avatarFileInput.value?.files?.[0]
            if (!file) return

            try {
                const arrayBuffer = await file.arrayBuffer()
                const byteArray = Array.from(new Uint8Array(arrayBuffer))
                const avatarURL = await UpdateAvatar(byteArray)
                emit('user-updated', {...props.user, avatarPath: avatarURL})
                showInfo('Аватар успешно обновлён')
            } catch (err) {
                showError(err)
            }

            if (avatarFileInput.value) avatarFileInput.value.value = ''
        }

        const deleteAvatar = async () => {
            isAvatarMenuOpen.value = false

            try {
                await DeleteAvatar()
                emit('user-updated', {...props.user, avatarPath: null})
                showInfo('Аватар удалён')
            } catch (err) {
                showError(err)
            }
        }

        onMounted(() => {
            loadSettings()
            document.addEventListener('click', closeAvatarMenu)
        })

        onUnmounted(() => {
            document.removeEventListener('click', closeAvatarMenu)
        })

        return {
            formatDate,
            isEditProfileOpen,
            editUsername,
            updateUser,
            isChangePasswordOpen,
            oldPassword,
            newPassword,
            confirmPassword,
            changePassword,
            isNotificationsOpen,
            emailNewMessageConsent,
            webPushNewMessageConsent,
            toggleEmailConsent,
            toggleWebPushConsent,
            isAvatarMenuOpen,
            avatarFileInput,
            toggleAvatarMenu,
            triggerFileInput,
            uploadAvatar,
            deleteAvatar,
        }
    },
}
