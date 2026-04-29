<template>
  <div class="chat-layout">
    <!-- Левая панель — список чатов -->
    <aside class="sidebar">
      <div class="sidebar__header">
        <h3>Чаты</h3>
        <button @click="$emit('show-create-chat')" class="sidebar__icon-btn" title="Создать чат">+</button>
        <button @click="$emit('show-search-users')" class="sidebar__icon-btn" title="Поиск пользователей">&#x1F50D;</button>
      </div>

      <div class="sidebar__list">
        <div
          v-for="chat in chats"
          :key="chat.id"
          :class="['chat-item', { 'chat-item--active': selectedChat?.id === chat.id, 'chat-item--unread': !chat.isRead }]"
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
          <div v-if="!chat.isRead" class="chat-item__unread-dot"></div>
        </div>
      </div>

      <div class="sidebar__footer">
        <label class="theme-switch">
          <span class="theme-switch__icon">{{ isDarkTheme ? '&#x1F319;' : '&#x2600;&#xFE0F;' }}</span>
          <div class="theme-switch__toggle" @click="toggleTheme">
            <div class="theme-switch__track" :class="{ 'theme-switch__track--on': isDarkTheme }">
              <div class="theme-switch__thumb" :class="{ 'theme-switch__thumb--on': isDarkTheme }"></div>
            </div>
          </div>
        </label>
        <button @click="handleLogout" class="sidebar__logout-btn">
          <span class="sidebar__logout-icon">&#x1F6AA;</span>
          <span>Выйти из аккаунта</span>
        </button>
      </div>
    </aside>

    <!-- Правая панель — сообщения -->
    <main class="conversation" v-if="selectedChat">
      <div class="conversation__header">
        <h3>{{ getChatTitle(selectedChat) }}</h3>
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
    </main>

    <div v-else class="conversation__placeholder">
      <p>Выберите чат для начала общения</p>
    </div>
  </div>
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>
