<template>
  <div class="modal-overlay"
       @click="$emit('close')"
       @keydown.escape="$emit('close')"
       tabindex="-1"
       v-focus
  >
    <div class="modal-content group-modal" @click.stop>
      <button class="modal-content__close" @click="$emit('close')" title="Закрыть">&times;</button>

      <div class="group-modal__header">
        <div class="group-modal__avatar">
          {{ chatInitial }}
        </div>
        <div class="group-modal__title">
          <h2>{{ chatTitle }}</h2>
          <span v-if="chat.description" class="group-modal__description">{{ chat.description }}</span>
        </div>
      </div>

      <div class="group-modal__members">
        <h4 class="group-modal__members-heading">Участники ({{ membersCount }})</h4>
        <div class="modal-content__users-list">
          <div
              v-for="member in chat.members"
              :key="member.id"
              class="user-item user-item--clickable"
              @click="openMemberProfile(member)"
          >
            <img
                v-if="member.avatarPath"
                class="user-item__avatar user-item__avatar--img"
                :src="member.avatarPath"
                :alt="member.username"
                @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''"
            />
            <div
                class="user-item__avatar"
                :style="member.avatarPath ? {display: 'none'} : {}"
            >
              {{ member.username.charAt(0).toUpperCase() }}
            </div>
            <div class="user-item__info">
              <div class="user-item__name">
                {{ member.username }}
                <span v-if="member.id === currentUser?.id" class="group-modal__you-badge">(вы)</span>
              </div>
              <div v-if="member.email" class="user-item__email">{{ member.email }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script src="./GroupChatModal.js"></script>
<style src="./GroupChatModal.css"></style>
