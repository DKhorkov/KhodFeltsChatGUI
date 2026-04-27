import {ref} from 'vue'
import {ForgetPassword} from '../../../wailsjs/go/forget_password/Handler'

export default {
    name: 'ForgetPasswordModal', emits: ['close'], props: ['forgetPasswordMessage'],

    setup(props, {emit}) {
        const token = ref('')
        const newPassword = ref('')
        const confirmPassword = ref('')

        const handleBack = () => {
            emit('close') // Отправляем событие закрытия
        }

        const resetPassword = async () => {
            if (!token.value || !newPassword.value || !confirmPassword.value) {
                alert(`Заполните все поля`)

                return
            }

            if (newPassword.value !== confirmPassword.value) {
                alert(`Пароли не совпадают`)

                return
            }

            try {
                await ForgetPassword(token.value, {
                    newPassword: newPassword.value
                })

                alert(`Пароль был успешно сброшен. Теперь вы можете авторизоваться.`)

                emit('close') // Закрываем окно после успеха
            } catch (err) {
                alert(err)
            }
        }

        return {
            token,
            newPassword,
            confirmPassword,
            resetPassword,
            handleBack,
        }
    }
}