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

      <div v-if="chatType === CHAT_TYPE.GROUP" class="modal-content__form-group">
        <label>Описание чата</label>
        <input v-model="chatDescription" placeholder="Описание чата" />
      </div>

      <div class="modal-content__form-group">
        <label>Поиск пользователей</label>
        <input
          v-model="searchQuery"
          @input="handleSearchInput"
          placeholder="Введите имя пользователя..."
        />
      </div>

      <div v-if="searchResults.length > 0" class="modal-content__users-list">
        <label
          v-for="user in searchResults"
          :key="user.id"
          class="user-item user-item--selectable"
        >
          <input
            type="checkbox"
            class="user-item__checkbox"
            :value="user.id"
            v-model="selectedUserIds"
          />
          <div class="user-item__avatar">
            {{ user.username.charAt(0).toUpperCase() }}
          </div>
          <div class="user-item__info">
            <div class="user-item__name">{{ user.username }}</div>
            <div class="user-item__email">{{ user.email }}</div>
          </div>
        </label>
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
