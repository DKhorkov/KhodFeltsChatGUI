import {inject, ref} from 'vue'
import {ForgetPassword} from '../../../wailsjs/go/forget_password/Handler'

export default {
    name: 'ForgetPasswordModal',
    emits: ['close'],
    props: {
        message: {
            type: String,
            required: true,
        },
    },

    setup(props, {emit}) {
        const showError = inject('showError')
        const showInfo = inject('showInfo')
        const token = ref('')
        const newPassword = ref('')
        const confirmPassword = ref('')

        const resetPassword = async () => {
            if (!token.value || !newPassword.value || !confirmPassword.value) {
                showInfo('Заполните все поля')
                return
            }

            if (newPassword.value !== confirmPassword.value) {
                showError('Пароли не совпадают')
                return
            }

            try {
                await ForgetPassword(token.value, {newPassword: newPassword.value})
                showInfo('Пароль был успешно сброшен. Теперь вы можете авторизоваться.')
                emit('close')
            } catch (err) {
                showError(err)
            }
        }

        return {
            token,
            newPassword,
            confirmPassword,
            resetPassword,
        }
    }
}
