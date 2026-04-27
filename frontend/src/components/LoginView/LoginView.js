import {ref} from 'vue'
import {Login, Register, SendForgetPassword, SendVerifyEmail} from '../../../wailsjs/go/auth/Handler'

export default {
    name: 'LoginView', emits: ['login-success', 'show-forget-password'],

    setup(props, {emit}) {
        const activeTab = ref('login')
        const loading = ref(false)

        const loginForm = ref({
            email: '', password: ''
        })

        const registerForm = ref({
            email: '', username: '', password: '', confirmPassword: ''
        })

        const handleLogin = async () => {
            loading.value = true

            try {
                await Login(loginForm.value.email, loginForm.value.password)

                emit('login-success')
            } catch (err) {
                alert(err)
            } finally {
                loading.value = false
            }
        }

        const handleRegister = async () => {
            if (registerForm.value.password !== registerForm.value.confirmPassword) {
                alert(`Пароли не совпадают`)

                return
            }

            loading.value = true

            try {
                await Register({
                    email: registerForm.value.email,
                    username: registerForm.value.username,
                    password: registerForm.value.password
                })

                activeTab.value = 'login'
                loginForm.value.email = registerForm.value.email
                loginForm.value.password = registerForm.value.password

                // Очищаем форму регистрации
                registerForm.value = {
                    email: '', username: '', password: '', confirmPassword: ''
                }

                alert(`Регистрация прошла успешно. Теперь войдите`)
            } catch (err) {
                alert(err)
            } finally {
                loading.value = false
            }
        }

        const sendVerifyEmail = async () => {
            if (!loginForm.value.email) {
                alert('Введите email')
                return
            }

            try {
                await SendVerifyEmail(loginForm.value.email)

                alert(`Письмо для подтверждения почты было отправлено по адресу ${loginForm.value.email}`)
            } catch (err) {
                alert(err)
            }
        }

        const sendForgetPassword = async () => {
            if (!loginForm.value.email) {
                alert('Введите email')
                return
            }

            try {
                await SendForgetPassword(loginForm.value.email)

                const forgetPasswordMessage = `Письмо с кодом для сброса пароля было отправлено по адресу ${loginForm.value.email}`

                emit('show-forget-password', forgetPasswordMessage)
            } catch (err) {
                alert(err)
            }
        }

        return {
            activeTab,
            loading,
            loginForm,
            registerForm,
            handleLogin,
            handleRegister,
            sendVerifyEmail,
            sendForgetPassword
        }
    }
}
