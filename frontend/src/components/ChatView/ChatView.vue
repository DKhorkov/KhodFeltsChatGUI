<template>
  <div class="chat-container">
    <!-- Левая панель - список чатов -->
    <div class="chats-panel">
      <div class="chats-panel__header">
        <h3>Чаты</h3>
        <button @click="$emit('show-create-chat')" class="chats-panel__icon-btn" title="Создать чат">+</button>
        <button @click="$emit('show-search-users')" class="chats-panel__icon-btn" title="Поиск пользователей">🔍</button>
      </div>

      <div class="chats-panel__list">
        <div
          v-for="chat in chats"
          :key="chat.id"
          :class="['chat-item', { 'chat-item--active': currentChat?.id === chat.id, 'chat-item--unread': !chat.isRead }]"
          @click="selectChat(chat)"
        >
          <div class="chat-item__avatar">
            {{ getChatTitle(chat).charAt(0).toUpperCase() }}
          </div>
          <div class="chat-item__info">
            <div class="chat-item__title" :class="{ 'chat-item__title--bold': !chat.isRead }">
              {{ getChatTitle(chat) }}
            </div>
          </div>
          <div v-if="!chat.isRead" class="chat-item__unread-indicator">●</div>
        </div>
      </div>

      <div class="chats-panel__footer">
        <label class="chats-panel__theme-toggle">
          <span class="chats-panel__theme-label">{{ isDarkTheme ? '🌙' : '☀️' }}</span>
          <div class="theme-toggle" @click="toggleTheme">
            <div class="theme-toggle__track" :class="{ 'theme-toggle__track--active': isDarkTheme }">
              <div class="theme-toggle__thumb" :class="{ 'theme-toggle__thumb--active': isDarkTheme }"></div>
            </div>
          </div>
        </label>
        <button @click="handleLogout" class="chats-panel__logout-btn">
          <span class="chats-panel__logout-icon">🚪</span>
          <span>Выйти из аккаунта</span>
        </button>
      </div>
    </div>

    <!-- Правая панель - сообщения -->
    <div class="messages-panel" v-if="currentChat">
      <div class="messages-panel__header">
        <h3>{{ getChatTitle(currentChat) }}</h3>
        <button @click="currentChat = null" class="messages-panel__close-btn" title="Закрыть чат">×</button>
      </div>

      <div class="messages-panel__list" ref="messagesList">
        <div
          v-for="message in messages"
          :key="message.id"
          :class="['message', { 'message--own': message.sender.id === currentUser?.id }]"
        >
          <div class="message__header">
            <span class="message__sender">{{ getSenderName(message) }}</span>
            <span class="message__time">{{ formatTime(message.createdAt) }}</span>
          </div>
          <div class="message__text">{{ message.text }}</div>
        </div>
      </div>

      <div class="messages-panel__input">
        <textarea
          v-model="newMessage"
          @keydown.enter.prevent="sendMessage"
          placeholder="Введите сообщение..."
          rows="3"
        ></textarea>
        <button @click="sendMessage" :disabled="!newMessage.trim()">
          Отправить
        </button>
      </div>
    </div>

    <div v-else class="chat-container__empty">
      <p>Выберите чат для начала общения</p>
    </div>
  </div>
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>