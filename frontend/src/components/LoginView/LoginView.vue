<template>
  <div class="login-container">
    <div class="login-card">
      <div class="tabs">
        <button
          :class="{ active: activeTab === 'login' }"
          @click="activeTab = 'login'"
        >
          Вход
        </button>
        <button
          :class="{ active: activeTab === 'register' }"
          @click="activeTab = 'register'"
        >
          Регистрация
        </button>
      </div>

      <div v-if="activeTab === 'login'" class="tab-content">
        <form @submit.prevent="handleLogin">
          <input
            v-model="loginForm.email"
            type="email"
            placeholder="Почта"
          />
          <input
            v-model="loginForm.password"
            type="password"
            placeholder="Пароль"
          />
          <button type="submit">Войти</button>
        </form>

        <div class="additional-buttons">
          <button @click="sendVerifyEmail" class="secondary">
            Отправить повторно письмо для подтверждения почты
          </button>
          <button @click="sendForgetPassword" class="danger">
            Сбросить пароль
          </button>
        </div>
      </div>

      <div v-else class="tab-content">
        <form @submit.prevent="handleRegister">
          <input
            v-model="registerForm.email"
            type="email"
            placeholder="Почта"
          />
          <input
            v-model="registerForm.username"
            type="text"
            placeholder="Логин"
          />
          <input
            v-model="registerForm.password"
            type="password"
            placeholder="Пароль"
          />
          <input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="Подтверждение пароля"
          />
          <button type="submit">Зарегистрироваться</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script src="./LoginView.js"></script>
<style scoped src="./LoginView.css"></style>
