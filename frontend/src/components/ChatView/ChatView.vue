<template>
  <div class="chat-layout">
    <!-- Верхний тулбар -->
    <div class="toolbar">
      <div class="toolbar__spacer"></div>
      <button @click="$emit('show-profile')" class="toolbar__profile-btn" v-if="currentUser" title="Профиль">
        <img
            v-if="currentUser.avatarPath"
            class="toolbar__profile-avatar toolbar__profile-avatar--img"
            :src="currentUser.avatarPath"
            :alt="currentUser.username"
            @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
        />
        <div
            class="toolbar__profile-avatar"
            :style="currentUser.avatarPath ? {display: 'none'} : {}"
        >
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
              :class="['chat-item', { 'chat-item--active': selectedChat?.id === chat.id, 'chat-item--unread': chat.unreadCount > 0 }]"
              @click="selectChat(chat)"
          >
            <img
                v-if="getChatAvatarPath(chat)"
                class="chat-item__avatar chat-item__avatar--img chat-item__avatar--clickable"
                :src="getChatAvatarPath(chat)"
                :alt="getChatTitle(chat)"
                @click.stop="openMemberProfile(chat)"
                @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
            />
            <div
                class="chat-item__avatar"
                :class="{ 'chat-item__avatar--clickable': true }"
                :style="getChatAvatarPath(chat) ? {display: 'none'} : {}"
                @click.stop="openMemberProfile(chat)"
            >
              {{ getChatTitle(chat).charAt(0).toUpperCase() }}
            </div>
            <div class="chat-item__info">
              <div class="chat-item__title" :class="{ 'chat-item__title--bold': chat.unreadCount > 0 }">
                {{ getChatTitle(chat) }}
              </div>
              <div v-if="getLastMessagePreview(chat)" class="chat-item__last-message">
                {{ getLastMessagePreview(chat) }}
              </div>
            </div>
            <div v-if="chat.unreadCount > 0" class="chat-item__unread-badge">
              {{ chat.unreadCount > 99 ? '99+' : chat.unreadCount }}
            </div>
          </div>
        </div>
      </aside>

      <!-- Правая панель — сообщения -->
      <main class="conversation" v-if="selectedChat" @keydown.escape="closeChat()" tabindex="-1" v-focus>
        <div class="conversation__header">
          <div class="conversation__header-left">
            <img
                v-if="getChatAvatarPath(selectedChat)"
                class="chat-item__avatar chat-item__avatar--img chat-item__avatar--clickable"
                :src="getChatAvatarPath(selectedChat)"
                :alt="getChatTitle(selectedChat)"
                @click.stop="openChatInfo"
                @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
            />
            <div
                class="chat-item__avatar chat-item__avatar--clickable"
                :style="getChatAvatarPath(selectedChat) ? {display: 'none'} : {}"
                @click.stop="openChatInfo"
            >
              {{ getChatTitle(selectedChat).charAt(0).toUpperCase() }}
            </div>
            <h3 class="conversation__header-title" @click="openChatInfo">{{ getChatTitle(selectedChat) }}</h3>
          </div>
          <button @click="closeChat()" class="conversation__close-btn" title="Закрыть чат">&times;</button>
        </div>

        <div class="conversation__messages" ref="messagesListRef">
          <template v-for="(message, index) in messages" :key="message.id">
            <div
                v-if="isFirstUnread(message, index)"
                class="conversation__unread-divider"
            >
              <span>Новые сообщения</span>
            </div>
            <div
                :class="['message-bubble', { 'message-bubble--own': message.sender.id === currentUser?.id, 'message-bubble--highlight': highlightedMessageId === message.id }]"
                :data-message-id="message.id"
                @contextmenu.prevent="openContextMenu($event, message)"
            >
              <div
                  v-if="message.replyToMessage"
                  class="message-bubble__reply"
                  @click="scrollToMessage(message.replyToMessage.id)"
              >
                <span class="message-bubble__reply-sender">
                  {{ message.replyToMessage.sender.id === currentUser?.id ? 'Вы' : message.replyToMessage.sender.username }}
                </span>
                <span class="message-bubble__reply-text">{{ message.replyToMessage.text }}</span>
              </div>
              <div class="message-bubble__header">
                <span class="message-bubble__sender">{{ getSenderName(message) }}</span>
                <span class="message-bubble__time">{{ formatTime(message.createdAt) }}</span>
              </div>
              <div class="message-bubble__text" v-html="linkifyToHtml(message.text)"></div>
              <div
                  v-if="Array.isArray(message.reactions) && message.reactions.length > 0"
                  class="message-bubble__reactions"
              >
                <button
                    v-for="summary in sortReactionsBySortOrder(message.reactions)"
                    :key="summary.reaction.id"
                    type="button"
                    :class="['message-bubble__reaction', { 'message-bubble__reaction--mine': isReactionSetForCurrentUser(message, summary.reaction.id) }]"
                    @click.stop="toggleReaction(message.id, summary.reaction.id)"
                >
                  <span class="message-bubble__reaction-emoji">{{ summary.reaction.emoji }}</span>
                  <span class="message-bubble__reaction-count">{{ summary.userIds.length }}</span>
                </button>
              </div>
            </div>
          </template>
        </div>

        <div class="conversation__composer">
          <button
              v-if="!isAtBottom"
              type="button"
              class="conversation__scroll-down"
              @click="onScrollDownClick"
              aria-label="К последнему сообщению"
          >
            <svg class="conversation__scroll-down-icon" viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span v-if="unreadMessageIds.size > 0" class="conversation__scroll-down-badge">
              {{ unreadMessageIds.size > 99 ? '99+' : unreadMessageIds.size }}
            </span>
          </button>
          <div v-if="editingMessage" class="conversation__edit-bar">
            <div class="conversation__edit-bar-content">
              <span class="conversation__edit-bar-label">Редактирование</span>
              <span class="conversation__edit-bar-text">{{ editingMessage.text }}</span>
            </div>
            <button class="conversation__edit-bar-close" @click="cancelEdit" title="Отменить редактирование">&times;</button>
          </div>
          <div v-else-if="replyToMessage" class="conversation__reply-bar">
            <div class="conversation__reply-bar-content">
              <span class="conversation__reply-bar-sender">
                {{ replyToMessage.sender.id === currentUser?.id ? 'Вы' : replyToMessage.sender.username }}
              </span>
              <span class="conversation__reply-bar-text">{{ replyToMessage.text }}</span>
            </div>
            <button class="conversation__reply-bar-close" @click="cancelReply" title="Отменить ответ">&times;</button>
          </div>
          <div class="conversation__composer-row">
            <div class="conversation__composer-input">
              <textarea
                  ref="textareaRef"
                  v-model="newMessage"
                  @keydown.enter.exact.prevent="sendMessage"
                  placeholder="Введите сообщение..."
                  rows="3"
              ></textarea>
              <div
                  class="conversation__emoji-wrapper"
                  @mouseenter="showEmojiPicker"
                  @mouseleave="scheduleEmojiClose"
              >
                <button
                    type="button"
                    class="conversation__emoji-toggle"
                    :class="{ 'conversation__emoji-toggle--active': isEmojiPickerVisible }"
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
              {{ editingMessage ? 'Сохранить' : 'Отправить' }}
            </button>
          </div>
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
          <img
              v-if="selectedMember.avatarPath"
              class="profile-modal__avatar profile-modal__avatar--img profile-modal__avatar--clickable"
              :src="selectedMember.avatarPath"
              :alt="selectedMember.username"
              @click="openAvatarZoom(selectedMember.avatarPath)"
              @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
          />
          <div
              class="profile-modal__avatar"
              :style="selectedMember.avatarPath ? {display: 'none'} : {}"
          >
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

    <!-- Увеличенный просмотр аватара -->
    <div
        v-if="avatarZoomSrc"
        class="avatar-zoom-overlay"
        @click.self="avatarZoomSrc = null"
        @keydown.escape.stop="avatarZoomSrc = null"
        tabindex="-1"
        v-focus
    >
      <button class="avatar-zoom-overlay__close" @click.stop="avatarZoomSrc = null" aria-label="Закрыть">&times;</button>
      <img class="avatar-zoom-overlay__img" :src="avatarZoomSrc" alt="Аватар" @click.stop />
    </div>

    <!-- Контекстное меню сообщения -->
    <div
        v-if="contextMenu.visible"
        class="context-menu"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
    >
      <button class="context-menu__item" @click="replyToContextMessage">Ответить</button>
      <button
          v-if="contextMenu.message?.sender.id === currentUser?.id"
          class="context-menu__item"
          @click="editContextMessage"
      >Редактировать</button>
      <button class="context-menu__item" @click="copyContextMessage">Копировать текст</button>
      <div v-if="contextMenu.message?.sender.id === currentUser?.id" class="context-menu__delete-group">
        <button
            v-if="!contextMenu.deleteExpanded"
            class="context-menu__item context-menu__item--danger"
            @click.stop="contextMenu.deleteExpanded = true"
        >Удалить</button>
        <template v-else>
          <button class="context-menu__item context-menu__item--danger" @click="deleteContextMessage(false)">Удалить у себя</button>
          <button class="context-menu__item context-menu__item--danger" @click="deleteContextMessage(true)">Удалить у всех</button>
        </template>
      </div>
      <div
          v-if="reactionsDictionary.length > 0"
          class="context-menu__reactions"
          @wheel.prevent="onReactionsBarWheel"
      >
        <button
            v-for="reaction in reactionsDictionary"
            :key="reaction.id"
            type="button"
            :class="['context-menu__reaction', { 'context-menu__reaction--active': isReactionSetForCurrentUser(contextMenu.message, reaction.id) }]"
            @click.stop="toggleReaction(contextMenu.message.id, reaction.id); contextMenu.visible = false"
        >{{ reaction.emoji }}</button>
      </div>
    </div>
  </div>
</template>

<script src="./ChatView.js"></script>
<style scoped src="./ChatView.css"></style>
