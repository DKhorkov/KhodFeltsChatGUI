<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2 class="modal-content__title">Сброс пароля</h2>

      <div class="modal-content__info">
        {{ forgetPasswordMessage }}
      </div>

      <div class="modal-content__form-group">
        <label>Код для сброса пароля</label>
        <input
          v-model="token"
          placeholder="Введите код из письма"
          @keydown.enter="resetPassword"
        />
      </div>

      <div class="modal-content__form-group">
        <label>Новый пароль</label>
        <input
          v-model="newPassword"
          type="password"
          placeholder="Новый пароль"
        />
      </div>

      <div class="modal-content__form-group">
        <label>Подтверждение пароля</label>
        <input
          v-model="confirmPassword"
          type="password"
          placeholder="Подтвердите пароль"
        />
      </div>

      <div class="modal-content__actions">
        <button @click="$emit('close')">Назад</button>
        <button @click="resetPassword">Сбросить пароль</button>
      </div>
    </div>
  </div>
</template>

<script src="./ForgetPasswordModal.js"></script>
<style scoped src="./ForgetPasswordModal.css"></style>
