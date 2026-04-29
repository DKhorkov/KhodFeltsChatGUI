<template>
  <div class="emoji-picker">
    <div class="emoji-picker__tabs">
      <button
        v-for="category in categories"
        :key="category.name"
        class="emoji-picker__tab"
        :class="{ 'emoji-picker__tab--active': activeCategory === category.name }"
        :title="category.label"
        @click="activeCategory = category.name"
      >
        {{ category.icon }}
      </button>
    </div>

    <div class="emoji-picker__grid">
      <button
        v-for="emoji in activeEmojis"
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
