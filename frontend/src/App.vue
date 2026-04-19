<template>
  <div id="app">
    <LoginView
      v-if="currentView === 'login'"
      @login-success="handleLoginSuccess"
      @show-forget-password="showForgetPasswordModal = true"
    />
    <ChatView
      v-else-if="currentView === 'chat'"
      @logout="handleLogout"
      @show-create-chat="showCreateChatModal = true"
      @show-search-users="showSearchUsersModal = true"
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
      @close="showForgetPasswordModal = false"
    />

    <NotificationToast
      v-if="notification"
      :message="notification"
      @close="notification = null"
    />
  </div>
</template>

<script src="./App.js"></script>
<style src="./App.css"></style>