<template>
  <div class="notification" @click="$emit('close')">
    <div class="notification-content">
      <div class="notification-icon">ℹ️</div>
      <div class="notification-message">{{ message }}</div>
      <button class="notification-close" @click="$emit('close')">×</button>
    </div>
  </div>
</template>

<script src="./NotificationToast.js"></script>
<style scoped src="./NotificationToast.css"></style>