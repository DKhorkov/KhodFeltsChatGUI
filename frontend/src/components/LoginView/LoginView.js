import {ref} from 'vue'
import {Login, Register, SendForgetPassword, SendVerifyEmail} from '../../../wailsjs/go/auth/Handler'
import {TAB} from '../../constants'

export default {
    name: 'LoginView', emits: ['login-success', 'show-forget-password'],

    setup(props, {emit}) {
        const activeTab = ref(TAB.LOGIN)

        const loginForm = ref({
            email: '', password: '',
        })

        const registerForm = ref({
            email: '', username: '', password: '', confirmPassword: '',
        })

        const handleLogin = async () => {
            if (!loginForm.value.email) {
                alert('Пожалуйста, введите адрес электронной почты')
                return
            }

            if (!loginForm.value.password) {
                alert('Пожалуйста, введите пароль')
                return
            }

            try {
                await Login(loginForm.value.email, loginForm.value.password)
                emit('login-success')
            } catch (err) {
                alert(err)
            }
        }

        const handleRegister = async () => {
            if (!registerForm.value.email) {
                alert('Пожалуйста, введите адрес электронной почты')
                return
            }

            if (!registerForm.value.username) {
                alert('Пожалуйста, введите логин')
                return
            }

            if (!registerForm.value.password) {
                alert('Пожалуйста, введите пароль')
                return
            }

            if (registerForm.value.password !== registerForm.value.confirmPassword) {
                alert('Пароли не совпадают')
                return
            }

            try {
                await Register({
                    email: registerForm.value.email,
                    username: registerForm.value.username,
                    password: registerForm.value.password,
                })

                activeTab.value = TAB.LOGIN
                loginForm.value.email = registerForm.value.email
                loginForm.value.password = registerForm.value.password

                registerForm.value = {email: '', username: '', password: '', confirmPassword: ''}

                alert('Регистрация прошла успешно. Теперь войдите')
            } catch (err) {
                alert(err)
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
                emit('show-forget-password', `Письмо с кодом для сброса пароля было отправлено по адресу ${loginForm.value.email}`)
            } catch (err) {
                alert(err)
            }
        }

        return {
            TAB,
            activeTab,
            loginForm,
            registerForm,
            handleLogin,
            handleRegister,
            sendVerifyEmail,
            sendForgetPassword,
        }
    }
}
