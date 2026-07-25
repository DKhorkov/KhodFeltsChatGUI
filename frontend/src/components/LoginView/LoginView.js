import {inject, ref} from 'vue'
import {Login, Register, SendForgetPassword, SendVerifyEmail} from '../../../wailsjs/go/auth/Handler'
import {TAB} from '../../constants'
import PasswordInput from '../PasswordInput/PasswordInput.vue'

export default {
    name: 'LoginView',
    components: {PasswordInput},
    emits: ['login-success', 'show-forget-password'],

    setup(props, {emit}) {
        const showError = inject('showError')
        const showInfo = inject('showInfo')
        const activeTab = ref(TAB.LOGIN)

        const loginForm = ref({
            login: '', password: '',
        })

        const registerForm = ref({
            email: '', username: '', password: '', confirmPassword: '',
        })

        const handleLogin = async () => {
            if (!loginForm.value.login) {
                showInfo('Пожалуйста, введите почту или имя пользователя')
                return
            }

            if (!loginForm.value.password) {
                showInfo('Пожалуйста, введите пароль')
                return
            }

            try {
                await Login({
                    login: loginForm.value.login,
                    password: loginForm.value.password,
                })
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

                const registeredEmail = registerForm.value.email

                activeTab.value = TAB.LOGIN
                loginForm.value.login = registerForm.value.email
                loginForm.value.password = registerForm.value.password

                registerForm.value = {email: '', username: '', password: '', confirmPassword: ''}

                showInfo(`Вы успешно зарегистрировались. Мы отправили письмо с подтверждением на ${registeredEmail} — перейдите по ссылке в письме, чтобы активировать аккаунт`)
            } catch (err) {
                showError(err)
            }
        }

        const sendVerifyEmail = async () => {
            if (!loginForm.value.login) {
                showInfo('Введите email в поле авторизации')
                return
            }

            try {
                await SendVerifyEmail(loginForm.value.login)
                showInfo(`Письмо для подтверждения почты было отправлено по адресу ${loginForm.value.login}`)
            } catch (err) {
                showError(err)
            }
        }

        const sendForgetPassword = async () => {
            if (!loginForm.value.login) {
                showInfo('Введите email в поле авторизации')
                return
            }

            try {
                await SendForgetPassword(loginForm.value.login)
                emit('show-forget-password', `Письмо с кодом для сброса пароля было отправлено по адресу ${loginForm.value.login}`)
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
