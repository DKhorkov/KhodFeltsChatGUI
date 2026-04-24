<template>
  <div class="chat-container">
    <!-- Левая панель - список чатов -->
    <div class="chats-panel">
      <div class="chats-header">
        <h3>Чаты</h3>
        <button @click="$emit('show-create-chat')" class="icon-btn">+</button>
        <button @click="$emit('show-search-users')" class="icon-btn">🔍</button>
        <button @click="handleLogout" class="icon-btn logout">🚪</button>
      </div>

      <div class="chats-list">
        <div
          v-for="chat in chats"
          :key="chat.id"
          :class="['chat-item', { active: currentChat?.id === chat.id, unread: !chat.isRead }]"
          @click="selectChat(chat)"
        >
          <div class="chat-avatar">
            {{ getChatTitle(chat).charAt(0).toUpperCase() }}
          </div>
          <div class="chat-info">
            <div class="chat-title" :class="{ bold: !chat.isRead }">
              {{ getChatTitle(chat) }}
            </div>
          </div>
          <div v-if="!chat.isRead" class="unread-indicator">●</div>
        </div>
      </div>
    </div>

    <!-- Правая панель - сообщения -->
    <div class="messages-panel" v-if="currentChat">
      <div class="messages-header">
        <h3>{{ getChatTitle(currentChat) }}</h3>
      </div>

      <div class="messages-list" ref="messagesList">
        <div
          v-for="message in messages"
          :key="message.id"
          :class="['message', { 'own': message.sender.id === currentUser?.id }]"
        >
          <div class="message-header">
            <span class="sender">{{ getSenderName(message) }}</span>
            <span class="time">{{ formatTime(message.createdAt) }}</span>
          </div>
          <div class="message-text">{{ message.text }}</div>
        </div>
      </div>

      <div class="message-input">
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

    <div v-else class="empty-panel">
      <p>Выберите чат для начала общения</p>
    </div>
  </div>
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>
