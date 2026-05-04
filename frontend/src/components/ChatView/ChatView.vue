<template>
  <div class="chat-layout">
    <!-- Верхний тулбар -->
    <div class="toolbar">
      <div class="toolbar__spacer"></div>
      <button @click="$emit('show-profile')" class="toolbar__profile-btn" v-if="currentUser" title="Профиль">
        <div class="toolbar__profile-avatar">
          {{ currentUser.username.charAt(0).toUpperCase() }}
        </div>
        <span class="toolbar__profile-name">{{ currentUser.username }}</span>
      </button>
    </div>

    <!-- Основная область: sidebar + conversation -->
    <div class="chat-layout__body">
      <!-- Левая панель — список чатов -->
      <aside class="sidebar">
        <div class="sidebar__header">
          <h3>Чаты</h3>
          <button @click="$emit('show-create-chat')" class="sidebar__icon-btn" title="Создать чат">+</button>
          <button @click="$emit('show-search-users')" class="sidebar__icon-btn" title="Поиск пользователей">&#x1F50D;
          </button>
        </div>

        <div class="sidebar__list">
          <div
              v-for="chat in chats"
              :key="chat.id"
              :class="['chat-item', { 'chat-item--active': selectedChat?.id === chat.id, 'chat-item--unread': !chat.isRead }]"
              @click="selectChat(chat)"
          >
            <div
                class="chat-item__avatar"
                :class="{ 'chat-item__avatar--clickable': true }"
                @click.stop="openMemberProfile(chat)"
            >
              {{ getChatTitle(chat).charAt(0).toUpperCase() }}
            </div>
            <div class="chat-item__info">
              <div class="chat-item__title" :class="{ 'chat-item__title--bold': !chat.isRead }">
                {{ getChatTitle(chat) }}
              </div>
            </div>
            <div v-if="!chat.isRead" class="chat-item__unread-dot"></div>
          </div>
        </div>
      </aside>

      <!-- Правая панель — сообщения -->
      <main class="conversation" v-if="selectedChat" @keydown.escape="selectedChat = null" tabindex="-1" v-focus>
        <div class="conversation__header">
          <h3 class="conversation__header-title" @click="openChatInfo">{{ getChatTitle(selectedChat) }}</h3>
          <button @click="selectedChat = null" class="conversation__close-btn" title="Закрыть чат">&times;</button>
        </div>

        <div class="conversation__messages" ref="messagesListRef">
          <template v-for="(message, index) in messages" :key="message.id">
            <div
                v-if="isFirstUnread(message, index)"
                class="conversation__unread-divider"
            >
              <span>Новые сообщения</span>
            </div>
            <div :class="['message-bubble', { 'message-bubble--own': message.sender.id === currentUser?.id }]">
              <div class="message-bubble__header">
                <span class="message-bubble__sender">{{ getSenderName(message) }}</span>
                <span class="message-bubble__time">{{ formatTime(message.createdAt) }}</span>
              </div>
              <div class="message-bubble__text">{{ message.text }}</div>
            </div>
          </template>
        </div>

        <div class="conversation__composer">
          <div class="conversation__composer-input">
            <textarea
                ref="textareaRef"
                v-model="newMessage"
                @keydown.enter.exact.prevent="sendMessage"
                placeholder="Введите сообщение..."
                rows="3"
            ></textarea>
            <div class="conversation__emoji-wrapper">
              <button
                  type="button"
                  class="conversation__emoji-toggle"
                  :class="{ 'conversation__emoji-toggle--active': isEmojiPickerVisible }"
                  @click="isEmojiPickerVisible = !isEmojiPickerVisible"
                  title="Смайлы"
              >
                &#x1F642;
              </button>
              <EmojiPicker
                  v-if="isEmojiPickerVisible"
                  @select="insertEmoji"
              />
            </div>
          </div>
          <button @click="sendMessage" :disabled="!newMessage.trim()">
            Отправить
          </button>
        </div>
      </main>

      <div v-else class="conversation__placeholder">
        <p>Выберите чат для начала общения</p>
      </div>
    </div>

    <!-- Профиль участника чата -->
    <div v-if="selectedMember"
         class="modal-overlay"
         @click="closeMemberProfile"
         @keydown.escape="closeMemberProfile"
         tabindex="-1"
         v-focus
    >
      <div class="modal-content profile-modal" @click.stop>
        <button class="modal-content__close" @click="closeMemberProfile" title="Закрыть">&times;</button>
        <div class="profile-modal__header">
          <div class="profile-modal__avatar">
            {{ selectedMember.username.charAt(0).toUpperCase() }}
          </div>
          <div class="profile-modal__title">
            <h2>{{ selectedMember.username }}</h2>
            <span class="profile-modal__email">{{ selectedMember.email }}</span>
          </div>
        </div>

        <div class="profile-modal__details">
          <div class="profile-modal__row">
            <span class="profile-modal__label">Email подтверждён</span>
            <span class="profile-modal__value"
                  :class="selectedMember.emailConfirmed ? 'profile-modal__value--success' : 'profile-modal__value--warning'">
              {{ selectedMember.emailConfirmed ? 'Да' : 'Нет' }}
            </span>
          </div>
          <div class="profile-modal__row">
            <span class="profile-modal__label">Дата регистрации</span>
            <span class="profile-modal__value">{{ formatDate(selectedMember.createdAt) }}</span>
          </div>
        </div>

      </div>
    </div>

    <!-- Информация о групповом чате -->
    <GroupChatModal
        v-if="selectedGroupChat"
        :chat="selectedGroupChat"
        :current-user="currentUser"
        @close="selectedGroupChat = null"
        @open-member-profile="openGroupMemberProfile"
    />
  </div>
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>
