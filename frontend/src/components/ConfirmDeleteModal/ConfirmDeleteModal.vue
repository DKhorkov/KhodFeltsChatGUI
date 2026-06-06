<template>
  <div class="modal-overlay modal-overlay--alert"
       @click.self="$emit('cancel')"
       @keydown.escape.stop="$emit('cancel')"
       tabindex="-1"
       v-focus
  >
    <div class="alert-modal" @click.stop>
      <div class="alert-modal__icon">❓</div>
      <div class="alert-modal__title">{{ title }}</div>
      <div class="alert-modal__message">{{ message }}</div>
      <div class="alert-modal__actions">
        <button
            class="alert-modal__btn alert-modal__btn--secondary"
            @click="$emit('cancel')"
        >{{ cancelText }}</button>
        <button
            class="alert-modal__btn"
            :class="`alert-modal__btn--${confirmType}`"
            @click="$emit('confirm')"
        >{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<script src="./ConfirmDeleteModal.js"></script>
<style src="./ConfirmDeleteModal.css"></style>
