import {inject, ref} from 'vue'
import {Login, Register, SendForgetPassword, SendVerifyEmail} from '../../../wailsjs/go/auth/Handler'
import {TAB} from '../../constants'

export default {
    name: 'LoginView', emits: ['login-success', 'show-forget-password'],

    setup(props, {emit}) {
        const showError = inject('showError')
        const showInfo = inject('showInfo')
        const activeTab = ref(TAB.LOGIN)

        const loginForm = ref({
            email: '', password: '',
        })

        const registerForm = ref({
            email: '', username: '', password: '', confirmPassword: '',
        })

        const handleLogin = async () => {
            if (!loginForm.value.email) {
                showInfo('Пожалуйста, введите адрес электронной почты')
                return
            }

            if (!loginForm.value.password) {
                showInfo('Пожалуйста, введите пароль')
                return
            }

            try {
                await Login(loginForm.value.email, loginForm.value.password)
                emit('login-success')
            } catch (err) {
                showError(err)
            }
        }

        const handleRegister = async () => {
            if (!registerForm.value.email) {
                showInfo('Пожалуйста, введите адрес электронной почты')
                return
            }

            if (!registerForm.value.username) {
                showInfo('Пожалуйста, введите логин')
                return
            }

            if (!registerForm.value.password) {
                showInfo('Пожалуйста, введите пароль')
                return
            }

            if (registerForm.value.password !== registerForm.value.confirmPassword) {
                showError('Пароли не совпадают')
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

                showInfo('Регистрация прошла успешно. Теперь войдите')
            } catch (err) {
                showError(err)
            }
        }

        const sendVerifyEmail = async () => {
            if (!loginForm.value.email) {
                showInfo('Введите email')
                return
            }

            try {
                await SendVerifyEmail(loginForm.value.email)
                showInfo(`Письмо для подтверждения почты было отправлено по адресу ${loginForm.value.email}`)
            } catch (err) {
                showError(err)
            }
        }

        const sendForgetPassword = async () => {
            if (!loginForm.value.email) {
                showInfo('Введите email')
                return
            }

            try {
                await SendForgetPassword(loginForm.value.email)
                emit('show-forget-password', `Письмо с кодом для сброса пароля было отправлено по адресу ${loginForm.value.email}`)
            } catch (err) {
                showError(err)
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
