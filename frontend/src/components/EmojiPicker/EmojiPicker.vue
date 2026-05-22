<template>
  <div class="emoji-picker">
    <div class="emoji-picker__grid">
      <button
        v-for="emoji in emojis"
        :key="emoji"
        class="emoji-picker__item"
        @click="$emit('select', emoji)"
      >
        {{ emoji }}
      </button>
    </div>
  </div>
</template>

<script src="./EmojiPicker.js"></script>
<style scoped src="./EmojiPicker.css"></style>
