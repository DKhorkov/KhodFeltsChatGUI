<template>
  <div class="modal-overlay modal-overlay--alert"
       @click="$emit('close')"
       @keydown.escape="$emit('close')"
       tabindex="-1"
       v-focus
  >
    <div class="alert-modal" :class="`alert-modal--${type}`" @click.stop>
      <div class="alert-modal__icon">{{ type === 'error' ? '❌' : 'ℹ️' }}</div>
      <div class="alert-modal__title">{{ type === 'error' ? 'Ошибка' : 'Информация' }}</div>
      <div class="alert-modal__message">{{ message }}</div>
      <button class="alert-modal__btn" :class="`alert-modal__btn--${type}`" @click="$emit('close')">OK</button>
    </div>
  </div>
</template>

<script src="./AlertModal.js"></script>
<style src="./AlertModal.css"></style>
