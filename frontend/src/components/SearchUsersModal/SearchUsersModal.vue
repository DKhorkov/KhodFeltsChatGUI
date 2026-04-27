<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2>Поиск пользователей</h2>

      <div class="form-group">
        <input
          v-model="searchQuery"
          @input="debouncedSearch"
          placeholder="Введите имя пользователя..."
        />
      </div>

      <div v-if="searchResults.length > 0" class="users-list">
        <div
          v-for="user in searchResults"
          :key="user.ID"
          class="user-item"
        >
          <div class="user-avatar">
            {{ user.username.charAt(0).toUpperCase() }}
          </div>
          <div class="user-info">
            <div class="username">{{ user.username }}</div>
            <div class="email">{{ user.email }}</div>
          </div>
        </div>
      </div>

      <div v-else-if="searched && searchResults.length === 0" class="no-results">
        Пользователи не найдены
      </div>

      <div class="modal-buttons">
        <button @click="$emit('close')">Закрыть</button>
      </div>
    </div>
  </div>
</template>

<script src="./SearchUsersModal.js"></script>
<style scoped src="./SearchUsersModal.css"></style>