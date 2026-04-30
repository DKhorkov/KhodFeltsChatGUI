import {inject, ref} from 'vue'
import {ChangePassword, UpdateUser} from '../../../wailsjs/go/profile/Handler'

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
        }
    },
}
