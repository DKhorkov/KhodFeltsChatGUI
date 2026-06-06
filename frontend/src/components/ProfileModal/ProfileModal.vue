<template>
  <div class="modal-overlay"
       @click.self="$emit('close')"
       @keydown.escape="onEscape"
       tabindex="-1"
       v-focus
  >
    <div class="modal-content profile-modal" @click.stop="closeAvatarMenu">
      <button class="modal-content__close" @click="$emit('close')" title="Закрыть">&times;</button>
      <div class="profile-modal__header">
        <div class="profile-modal__avatar-wrapper" @click.stop="toggleAvatarMenu">
          <img
              v-if="user.avatarPath"
              class="profile-modal__avatar profile-modal__avatar--img"
              :src="user.avatarPath"
              :alt="user.username"
              @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
          />
          <div
              class="profile-modal__avatar"
              :style="user.avatarPath ? {display: 'none'} : {}"
          >
            {{ user.username.charAt(0).toUpperCase() }}
          </div>
          <div class="avatar-context-menu" v-if="isAvatarMenuOpen">
            <button
                v-if="user.avatarPath"
                class="avatar-context-menu__item"
                @click.stop="openAvatarZoom"
            >Открыть фото</button>
            <button class="avatar-context-menu__item" @click.stop="triggerFileInput">Изменить фото</button>
            <button
                v-if="user.avatarPath"
                class="avatar-context-menu__item avatar-context-menu__item--danger"
                @click.stop="askDeleteAvatar"
            >Удалить фото</button>
          </div>
        </div>
        <input type="file" ref="avatarFileInput" accept="image/*" style="display: none;" @change="uploadAvatar" />
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
            class="profile-modal__toggle"
            @click="isEditProfileOpen = !isEditProfileOpen"
        >
          <span class="profile-modal__label">Редактировать профиль</span>
          <span class="profile-modal__chevron"
                :class="{ 'profile-modal__chevron--open': isEditProfileOpen }">&#9654;</span>
        </div>

        <form
            v-if="isEditProfileOpen"
            class="profile-modal__form"
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
            class="profile-modal__toggle"
            @click="isChangePasswordOpen = !isChangePasswordOpen"
        >
          <span class="profile-modal__label">Сменить пароль</span>
          <span class="profile-modal__chevron"
                :class="{ 'profile-modal__chevron--open': isChangePasswordOpen }">&#9654;</span>
        </div>

        <form
            v-if="isChangePasswordOpen"
            class="profile-modal__form"
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
        <div
            class="profile-modal__toggle"
            @click="isNotificationsOpen = !isNotificationsOpen"
        >
          <span class="profile-modal__label">Уведомления</span>
          <span class="profile-modal__chevron"
                :class="{ 'profile-modal__chevron--open': isNotificationsOpen }">&#9654;</span>
        </div>

        <div v-if="isNotificationsOpen" class="profile-modal__notifications">
          <div class="notifications__group">
            <span class="notifications__group-label">Новые сообщения</span>
            <div class="notifications__toggle-row">
              <span class="notifications__channel-label">Email</span>
              <div class="toggle" @click="toggleEmailConsent">
                <div class="toggle__track" :class="{ 'toggle__track--on': emailNewMessageConsent }">
                  <div class="toggle__thumb" :class="{ 'toggle__thumb--on': emailNewMessageConsent }"></div>
                </div>
              </div>
            </div>
            <div class="notifications__toggle-row">
              <span class="notifications__channel-label">Web Push</span>
              <div class="toggle" @click="toggleWebPushConsent">
                <div class="toggle__track" :class="{ 'toggle__track--on': webPushNewMessageConsent }">
                  <div class="toggle__thumb" :class="{ 'toggle__thumb--on': webPushNewMessageConsent }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="profile-modal__section">
        <label class="profile-modal__theme-row" @click="$emit('toggle-theme')">
          <span class="profile-modal__label">Тёмная тема</span>
          <div class="toggle">
            <div class="toggle__track" :class="{ 'toggle__track--on': isDarkTheme }">
              <div class="toggle__thumb" :class="{ 'toggle__thumb--on': isDarkTheme }"></div>
            </div>
          </div>
        </label>
      </div>

      <div class="modal-content__actions">
        <button class="btn--danger" @click="$emit('logout')">Выйти из аккаунта</button>
      </div>
    </div>

    <div
        v-if="isAvatarZoomOpen"
        class="avatar-zoom-overlay"
        @click.stop="closeAvatarZoom"
        @keydown.escape.stop="closeAvatarZoom"
        tabindex="-1"
        v-focus
    >
      <button class="avatar-zoom-overlay__close" @click.stop="closeAvatarZoom" aria-label="Закрыть">&times;</button>
      <img class="avatar-zoom-overlay__img" :src="user.avatarPath" alt="Аватар" @click.stop />
    </div>

    <ConfirmDeleteModal
        v-if="isConfirmDeleteAvatarOpen"
        message="Вы уверены, что хотите удалить фото профиля?"
        @confirm="confirmDeleteAvatar"
        @cancel="isConfirmDeleteAvatarOpen = false"
    />
  </div>
</template>

<script src="./ProfileModal.js"></script>
<style src="./ProfileModal.css"></style>
