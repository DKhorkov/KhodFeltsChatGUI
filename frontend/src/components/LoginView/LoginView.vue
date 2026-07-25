<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-card__tabs">
        <button
          class="login-card__tab"
          :class="{ 'login-card__tab--active': activeTab === TAB.LOGIN }"
          @click="activeTab = TAB.LOGIN"
        >
          Вход
        </button>
        <button
          class="login-card__tab"
          :class="{ 'login-card__tab--active': activeTab === TAB.REGISTER }"
          @click="activeTab = TAB.REGISTER"
        >
          Регистрация
        </button>
      </div>

      <div v-if="activeTab === TAB.LOGIN" class="login-card__tab-content">
        <form @submit.prevent="handleLogin" class="login-form">
          <input
            v-model="loginForm.login"
            class="login-form__input"
            type="text"
            placeholder="Почта или имя пользователя"
          />
          <PasswordInput
            v-model="loginForm.password"
            input-class="login-form__input"
            placeholder="Пароль"
          />
          <button type="submit" class="login-form__submit">Войти</button>
        </form>

        <div class="login-form__links">
          <button @click="sendVerifyEmail" class="login-form__link" type="button">
            Отправить повторно письмо для подтверждения почты
          </button>
          <button @click="sendForgetPassword" class="login-form__link login-form__link--danger" type="button">
            Сбросить пароль
          </button>
        </div>
      </div>

      <div v-else class="login-card__tab-content">
        <form @submit.prevent="handleRegister" class="login-form">
          <input
            v-model="registerForm.email"
            class="login-form__input"
            type="email"
            placeholder="Почта"
          />
          <input
            v-model="registerForm.username"
            class="login-form__input"
            type="text"
            placeholder="Логин"
          />
          <PasswordInput
            v-model="registerForm.password"
            input-class="login-form__input"
            placeholder="Пароль"
            autocomplete="new-password"
          />
          <PasswordInput
            v-model="registerForm.confirmPassword"
            input-class="login-form__input"
            placeholder="Подтверждение пароля"
            autocomplete="new-password"
          />
          <button type="submit" class="login-form__submit">Зарегистрироваться</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script src="./LoginView.js"></script>
<style scoped src="./LoginView.css"></style>
