<template>
  <div id="app">
    <LoginView
      v-if="currentView === VIEW.LOGIN"
      @login-success="handleLoginSuccess"
      @show-forget-password="handleShowForgetPassword"
    />
    <ChatView
      ref="chatViewComponent"
      v-else-if="currentView === VIEW.CHAT"
      @logout="handleLogout"
      @show-create-chat="showCreateChatModal = true"
      @show-search-users="showSearchUsersModal = true"
      @new-message-notification="handleNewMessageNotification"
    />

    <!-- Модальные окна -->
    <CreateChatModal
      v-if="showCreateChatModal"
      @close="showCreateChatModal = false"
      @chat-created="handleChatCreated"
    />

    <SearchUsersModal
      v-if="showSearchUsersModal"
      @close="showSearchUsersModal = false"
    />

    <ForgetPasswordModal
      v-if="showForgetPasswordModal"
      :forgetPasswordMessage="forgetPasswordMessage"
      @close="showForgetPasswordModal = false"
    />

    <NotificationToast
      v-if="notification"
      :message="notification"
      @click="handleNotificationClick"
      @close="notification = null"
    />
  </div>
</template>

<script src="./App.js"></script>
<style src="./App.css"></style>