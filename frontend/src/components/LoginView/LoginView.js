import { ref } from 'vue'

export default {
    name: 'LoginView',
    emits: ['login-success', 'show-forget-password'],

    setup(props, { emit }) {
        const activeTab = ref('login')
        const loading = ref(false)
        const error = ref('')
        const success = ref('')

        const loginForm = ref({
            email: '',
            password: ''
        })

        const registerForm = ref({
            email: '',
            username: '',
            password: '',
            confirmPassword: ''
        })

        const handleLogin = async () => {
            loading.value = true
            error.value = ''

            try {
                const result = await window.go.main.AuthHandler.Login(loginForm.value)
                if (result.Success) {
                    emit('login-success')
                } else {
                    error.value = result.Message
                }
            } catch (err) {
                error.value = err.message
            } finally {
                loading.value = false
            }
        }

        const handleRegister = async () => {
            if (registerForm.value.password !== registerForm.value.confirmPassword) {
                error.value = 'Пароли не совпадают'
                return
            }

            loading.value = true
            error.value = ''

            try {
                const result = await window.go.main.AuthHandler.Register({
                    email: registerForm.value.email,
                    username: registerForm.value.username,
                    password: registerForm.value.password
                })

                if (result.Success) {
                    success.value = result.Message
                    activeTab.value = 'login'
                    loginForm.value.email = registerForm.value.email
                    loginForm.value.password = registerForm.value.password

                    // Очищаем форму регистрации
                    registerForm.value = {
                        email: '',
                        username: '',
                        password: '',
                        confirmPassword: ''
                    }
                } else {
                    error.value = result.Message
                }
            } catch (err) {
                error.value = err.message
            } finally {
                loading.value = false
            }
        }

        const sendVerifyEmail = async () => {
            if (!loginForm.value.email) {
                error.value = 'Введите email'
                return
            }

            try {
                await window.go.main.AuthHandler.SendVerifyEmail(loginForm.value.email)
                success.value = `Письмо для подтверждения почты было отправлено по адресу ${loginForm.value.email}`
                setTimeout(() => {
                    success.value = ''
                }, 5000)
            } catch (err) {
                error.value = err.message
            }
        }

        return {
            activeTab,
            loading,
            error,
            success,
            loginForm,
            registerForm,
            handleLogin,
            handleRegister,
            sendVerifyEmail
        }
    }
}
