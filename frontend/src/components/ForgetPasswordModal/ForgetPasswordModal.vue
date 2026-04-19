<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2>Сброс пароля</h2>

      <div v-if="!tokenSent" class="step">
        <div class="form-group">
          <label>Email</label>
          <input
            v-model="email"
            type="email"
            placeholder="Введите ваш email"
            @keydown.enter="sendResetCode"
          />
        </div>

        <div class="modal-buttons">
          <button @click="$emit('close')">Отмена</button>
          <button @click="sendResetCode" :disabled="loading">
            {{ loading ? 'Отправка...' : 'Отправить код' }}
          </button>
        </div>
      </div>

      <div v-else class="step">
        <div class="info-message">
          {{ message }}
        </div>

        <div class="form-group">
          <label>Код для сброса пароля</label>
          <input
            v-model="token"
            placeholder="Введите код из письма"
            @keydown.enter="resetPassword"
          />
        </div>

        <div class="form-group">
          <label>Новый пароль</label>
          <input
            v-model="newPassword"
            type="password"
            placeholder="Новый пароль"
          />
        </div>

        <div class="form-group">
          <label>Подтверждение пароля</label>
          <input
            v-model="confirmPassword"
            type="password"
            placeholder="Подтвердите пароль"
          />
        </div>

        <div class="modal-buttons">
          <button @click="tokenSent = false">Назад</button>
          <button @click="resetPassword" :disabled="loading">
            {{ loading ? 'Сброс...' : 'Сбросить пароль' }}
          </button>
        </div>
      </div>

      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="success" class="success">{{ success }}</div>
    </div>
  </div>
</template>

<script src="./ForgetPasswordModal.js"></script>
<style scoped src="./ForgetPasswordModal.css"></style>