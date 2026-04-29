<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content profile-modal" @click.stop>
      <button class="profile-modal__close" @click="$emit('close')" title="Закрыть">&times;</button>
      <div class="profile-modal__header">
        <div class="profile-modal__avatar">
          {{ user.username.charAt(0).toUpperCase() }}
        </div>
        <div class="profile-modal__title">
          <h2>{{ user.username }}</h2>
          <span class="profile-modal__email">{{ user.email }}</span>
        </div>
      </div>

      <div class="profile-modal__details">
        <div class="profile-modal__row">
          <span class="profile-modal__label">Email подтверждён</span>
          <span class="profile-modal__value" :class="user.emailConfirmed ? 'profile-modal__value--success' : 'profile-modal__value--warning'">
            {{ user.emailConfirmed ? 'Да' : 'Нет' }}
          </span>
        </div>
        <div class="profile-modal__row">
          <span class="profile-modal__label">Дата регистрации</span>
          <span class="profile-modal__value">{{ formatDate(user.createdAt) }}</span>
        </div>
      </div>

      <div class="profile-modal__section">
        <label class="profile-modal__theme-row">
          <span class="profile-modal__label">Тёмная тема</span>
          <div class="theme-switch__toggle" @click="$emit('toggle-theme')">
            <div class="theme-switch__track" :class="{ 'theme-switch__track--on': isDarkTheme }">
              <div class="theme-switch__thumb" :class="{ 'theme-switch__thumb--on': isDarkTheme }"></div>
            </div>
          </div>
        </label>
      </div>

      <div class="modal-content__actions">
        <button class="btn--danger" @click="$emit('logout')">Выйти из аккаунта</button>
      </div>
    </div>
  </div>
</template>

<script src="./ProfileModal.js"></script>
<style src="./ProfileModal.css"></style>
