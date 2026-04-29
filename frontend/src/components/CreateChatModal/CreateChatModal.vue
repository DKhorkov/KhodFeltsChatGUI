<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2 class="modal-content__title">Создать чат</h2>

      <div class="modal-content__form-group">
        <label>Тип чата</label>
        <select v-model="chatType">
          <option :value="CHAT_TYPE.PRIVATE">Приватный</option>
          <option :value="CHAT_TYPE.GROUP">Групповой</option>
        </select>
      </div>

      <div v-if="chatType === CHAT_TYPE.GROUP" class="modal-content__form-group">
        <label>Название чата</label>
        <input v-model="chatTitle" placeholder="Название чата" />
      </div>

      <div class="modal-content__form-group">
        <label>Поиск пользователей</label>
        <input
          v-model="searchQuery"
          @input="debouncedSearch"
          placeholder="Введите имя пользователя..."
        />
      </div>

      <div v-if="searchResults.length > 0" class="modal-content__users-list">
        <div
          v-for="user in searchResults"
          :key="user.id"
          class="user-item"
        >
          <input
            type="checkbox"
            :value="user.id"
            v-model="selectedUsers"
          />
          <div class="user-item__info">
            <div class="user-item__name">{{ user.username }}</div>
            <div class="user-item__email">{{ user.email }}</div>
          </div>
        </div>
      </div>

      <div class="modal-content__actions">
        <button class="btn--secondary" @click="$emit('close')">Отмена</button>
        <button class="btn--primary" @click="createChat">Создать</button>
      </div>
    </div>
  </div>
</template>

<script src="./CreateChatModal.js"></script>
<style src="./CreateChatModal.css"></style>
