import { ref } from 'vue'

export default {
  name: 'ForgetPasswordModal',
  emits: ['close'],

  setup(props, { emit }) {
    const email = ref('')
    const token = ref('')
    const newPassword = ref('')
    const confirmPassword = ref('')
    const tokenSent = ref(false)
    const message = ref('')
    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const sendResetCode = async () => {
      if (!email.value) {
        error.value = 'Введите email'
        return
      }

      loading.value = true
      error.value = ''

      try {
        await window.go.main.AuthHandler.SendForgetPassword(email.value)
        message.value = `Письмо с кодом для сброса пароля было отправлено по адресу ${email.value}`
        tokenSent.value = true
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }

    const resetPassword = async () => {
      if (!token.value || !newPassword.value || !confirmPassword.value) {
        error.value = 'Заполните все поля'
        return
      }

      if (newPassword.value !== confirmPassword.value) {
        error.value = 'Пароли не совпадают'
        return
      }

      if (newPassword.value.length < 6) {
        error.value = 'Пароль должен содержать минимум 6 символов'
        return
      }

      loading.value = true
      error.value = ''

      try {
        await window.go.main.ForgetPasswordHandler.ResetPassword({
          token: token.value,
          newPassword: newPassword.value
        })

        success.value = 'Пароль был успешно сброшен. Теперь вы можете авторизоваться.'
        setTimeout(() => {
          emit('close')
        }, 2000)
      } catch (err) {
        error.value = err.message
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
      error,
      success,
      sendResetCode,
      resetPassword
    }
  }
}