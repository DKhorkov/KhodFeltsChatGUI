<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2>Создать чат</h2>

      <div class="form-group">
        <label>Тип чата</label>
        <select v-model="chatType">
          <option value="private">Приватный</option>
          <option value="group">Групповой</option>
        </select>
      </div>

      <div v-if="chatType === 'group'" class="form-group">
        <label>Название чата</label>
        <input v-model="chatTitle" placeholder="Название чата" />
      </div>

      <div class="form-group">
        <label>Поиск пользователей</label>
        <input
          v-model="searchQuery"
          @input="debouncedSearch"
          placeholder="Введите имя пользователя..."
        />
      </div>

      <div v-if="searchResults.length > 0" class="users-list">
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
          <div class="user-info">
            <div class="username">{{ user.username }}</div>
            <div class="email">{{ user.email }}</div>
          </div>
        </div>
      </div>

      <div class="modal-buttons">
        <button @click="$emit('close')">Отмена</button>
        <button @click="createChat">Создать</button>
      </div>
    </div>
  </div>
</template>

<script src="./CreateChatModal.js"></script>
<style scoped src="./CreateChatModal.css"></style>