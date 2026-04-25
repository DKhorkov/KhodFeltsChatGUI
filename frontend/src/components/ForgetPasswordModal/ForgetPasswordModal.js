import {ref} from 'vue'
import {SendForgetPassword} from '../../../wailsjs/go/auth/Handler'
import {ForgetPassword} from '../../../wailsjs/go/forget_password/Handler'

export default {
    name: 'ForgetPasswordModal', emits: ['close'],

    setup(props, {emit}) {
        const email = ref('')
        const token = ref('')
        const newPassword = ref('')
        const confirmPassword = ref('')
        const tokenSent = ref(false)
        const message = ref('')
        const loading = ref(false)

        const sendResetCode = async () => {
            if (!email.value) {
                alert(`Введите email`)

                return
            }

            loading.value = true

            try {
                await SendForgetPassword(email.value)

                message.value = `Письмо с кодом для сброса пароля было отправлено по адресу ${email.value}`

                tokenSent.value = true
            } catch (err) {
                alert(err)
            } finally {
                loading.value = false
            }
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

            loading.value = true

            try {
                await ForgetPassword(token.value, {
                    newPassword: newPassword.value
                })

                alert(`Пароль был успешно сброшен. Теперь вы можете авторизоваться.`)
            } catch (err) {
                alert(err)
            } finally {
                loading.value = false
            }
        }

        return {
            email,
            token,
            newPassword,
            confirmPassword,
            tokenSent,
            message,
            loading,
            sendResetCode,
            resetPassword
        }
    }
}