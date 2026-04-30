<template>
  <div class="modal-overlay"
       @click="$emit('close')"
       @keydown.escape="$emit('close')"
       tabindex="-1"
       v-focus
  >
    <div class="modal-content profile-modal" @click.stop>
      <button class="modal-content__close" @click="$emit('close')" title="Закрыть">&times;</button>
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
          <span class="profile-modal__value"
                :class="user.emailConfirmed ? 'profile-modal__value--success' : 'profile-modal__value--warning'">
            {{ user.emailConfirmed ? 'Да' : 'Нет' }}
          </span>
        </div>
        <div class="profile-modal__row">
          <span class="profile-modal__label">Дата регистрации</span>
          <span class="profile-modal__value">{{ formatDate(user.createdAt) }}</span>
        </div>
      </div>

      <div class="profile-modal__section">
        <div
            class="profile-modal__change-password-toggle"
            @click="isEditProfileOpen = !isEditProfileOpen"
        >
          <span class="profile-modal__label">Редактировать профиль</span>
          <span class="profile-modal__chevron"
                :class="{ 'profile-modal__chevron--open': isEditProfileOpen }">&#9654;</span>
        </div>

        <form
            v-if="isEditProfileOpen"
            class="profile-modal__change-password-form"
            @submit.prevent="updateUser"
        >
          <div class="modal-content__form-group">
            <label>Логин</label>
            <input
                v-model="editUsername"
                type="text"
                placeholder="Введите новый логин"
            />
          </div>
          <div class="modal-content__actions">
            <button type="submit" class="btn--primary">Сохранить</button>
          </div>
        </form>
      </div>

      <div class="profile-modal__section">
        <div
            class="profile-modal__change-password-toggle"
            @click="isChangePasswordOpen = !isChangePasswordOpen"
        >
          <span class="profile-modal__label">Сменить пароль</span>
          <span class="profile-modal__chevron"
                :class="{ 'profile-modal__chevron--open': isChangePasswordOpen }">&#9654;</span>
        </div>

        <form
            v-if="isChangePasswordOpen"
            class="profile-modal__change-password-form"
            @submit.prevent="changePassword"
        >
          <div class="modal-content__form-group">
            <label>Старый пароль</label>
            <input
                v-model="oldPassword"
                type="password"
                placeholder="Введите старый пароль"
            />
          </div>
          <div class="modal-content__form-group">
            <label>Новый пароль</label>
            <input
                v-model="newPassword"
                type="password"
                placeholder="Введите новый пароль"
            />
          </div>
          <div class="modal-content__form-group">
            <label>Подтверждение пароля</label>
            <input
                v-model="confirmPassword"
                type="password"
                placeholder="Подтвердите новый пароль"
            />
          </div>
          <div class="modal-content__actions">
            <button type="submit" class="btn--primary">Сменить пароль</button>
          </div>
        </form>
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
