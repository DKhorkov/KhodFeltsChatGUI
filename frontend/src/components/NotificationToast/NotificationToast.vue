<template>
  <div class="notification" @click="$emit('close')">
    <div class="notification__content">
      <div class="notification__icon">ℹ️</div>
      <div class="notification__message">{{ message }}</div>
      <button class="notification__close" @click.stop="$emit('close')">×</button>
    </div>
  </div>
</template>

<script src="./NotificationToast.js"></script>
<style scoped src="./NotificationToast.css"></style>
