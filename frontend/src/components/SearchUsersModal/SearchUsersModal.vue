<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <h2 class="modal-content__title">Поиск пользователей</h2>

      <div class="modal-content__form-group">
        <input
          v-model="searchQuery"
          @input="handleSearchInput"
          placeholder="Введите имя пользователя..."
        />
      </div>

      <div v-if="searchResults.length > 0" class="modal-content__users-list">
        <div
          v-for="user in searchResults"
          :key="user.id"
          class="user-item"
        >
          <div class="user-item__avatar">
            {{ user.username.charAt(0).toUpperCase() }}
          </div>
          <div class="user-item__info">
            <div class="user-item__name">{{ user.username }}</div>
            <div class="user-item__email">{{ user.email }}</div>
          </div>
        </div>
      </div>

      <div v-else-if="hasSearched && searchResults.length === 0" class="modal-content__no-results">
        Пользователи не найдены
      </div>

      <div class="modal-content__actions">
        <button class="btn--secondary" @click="$emit('close')">Закрыть</button>
      </div>
    </div>
  </div>
</template>

<script src="./SearchUsersModal.js"></script>
<style src="./SearchUsersModal.css"></style>
