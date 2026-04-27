<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2>Сброс пароля</h2>

      <div class="step">
        <div class="info-message">
          {{ forgetPasswordMessage }}
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
          <button @click="handleBack">Назад</button>
          <button @click="resetPassword">Сбросить пароль</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script src="./ForgetPasswordModal.js"></script>
<style scoped src="./ForgetPasswordModal.css"></style>