<template>
  <div class="toast" @click="$emit('click')">
    <div class="toast__body">
      <div v-if="sender" class="toast__sender">{{ sender }}</div>
      <div class="toast__text">{{ message }}</div>
    </div>
    <button class="toast__close" @click.stop="$emit('close')">×</button>
  </div>
</template>

<script src="./NotificationToast.js"></script>
<style scoped src="./NotificationToast.css"></style>
