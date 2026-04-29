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
        <button @click="$emit('show-profile')" class="sidebar__profile-btn" v-if="currentUser">
          <div class="sidebar__profile-avatar">
            {{ currentUser.username.charAt(0).toUpperCase() }}
          </div>
          <span class="sidebar__profile-name">{{ currentUser.username }}</span>
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
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>
