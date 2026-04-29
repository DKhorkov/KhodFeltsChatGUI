<template>
  <div id="app">
    <LoginView
      v-if="currentView === VIEW.LOGIN"
      @login-success="handleLoginSuccess"
      @show-forget-password="handleShowForgetPassword"
    />
    <ChatView
      ref="chatViewRef"
      v-else-if="currentView === VIEW.CHAT"
      @logout="handleLogout"
      @show-create-chat="isCreateChatVisible = true"
      @show-search-users="isSearchUsersVisible = true"
      @new-message-notification="handleNewMessageNotification"
    />

    <CreateChatModal
      v-if="isCreateChatVisible"
      @close="isCreateChatVisible = false"
      @chat-created="handleChatCreated"
    />

    <SearchUsersModal
      v-if="isSearchUsersVisible"
      @close="isSearchUsersVisible = false"
    />

    <ForgetPasswordModal
      v-if="isForgetPasswordVisible"
      :message="forgetPasswordMessage"
      @close="isForgetPasswordVisible = false"
    />

    <AlertModal
      v-if="alert"
      :message="alert.message"
      :type="alert.type"
      @close="alert = null"
    />

    <div class="notifications-stack">
      <NotificationToast
        v-for="notification in notifications"
        :key="notification.id"
        :message="notification.message"
        @click="handleNotificationClick(notification.id)"
        @close="removeNotification(notification.id)"
      />
    </div>
  </div>
</template>

<script src="./App.js"></script>
<style src="./App.css"></style>
