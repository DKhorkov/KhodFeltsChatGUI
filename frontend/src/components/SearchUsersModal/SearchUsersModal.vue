<template>
  <div class="modal-overlay"
       @click="$emit('close')"
       @keydown.escape="$emit('close')"
       tabindex="-1"
       v-focus
  >
    <div class="modal-content" @click.stop v-if="!selectedUser">
      <button class="modal-content__close" @click="$emit('close')" title="Закрыть">&times;</button>
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
            class="user-item user-item--clickable"
            @click="selectedUser = user"
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

    </div>

    <!-- Профиль выбранного пользователя -->
    <div class="modal-content profile-modal" @click.stop v-else>
      <button class="modal-content__close" @click="selectedUser = null" title="Назад">&times;</button>
      <div class="profile-modal__header">
        <div class="profile-modal__avatar">
          {{ selectedUser.username.charAt(0).toUpperCase() }}
        </div>
        <div class="profile-modal__title">
          <h2>{{ selectedUser.username }}</h2>
          <span class="profile-modal__email">{{ selectedUser.email }}</span>
        </div>
      </div>

      <div class="profile-modal__details">
        <div class="profile-modal__row">
          <span class="profile-modal__label">Email подтверждён</span>
          <span class="profile-modal__value"
                :class="selectedUser.emailConfirmed ? 'profile-modal__value--success' : 'profile-modal__value--warning'">
            {{ selectedUser.emailConfirmed ? 'Да' : 'Нет' }}
          </span>
        </div>
        <div class="profile-modal__row">
          <span class="profile-modal__label">Дата регистрации</span>
          <span class="profile-modal__value">{{ formatDate(selectedUser.createdAt) }}</span>
        </div>
      </div>

    </div>
  </div>
</template>

<script src="./SearchUsersModal.js"></script>
<style src="./SearchUsersModal.css"></style>
