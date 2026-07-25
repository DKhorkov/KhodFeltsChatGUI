<template>
  <div class="password-field">
    <input
      :id="id"
      :name="name"
      :type="visible ? 'text' : 'password'"
      :class="['password-field__input', inputClass]"
      :placeholder="placeholder"
      :required="required"
      :autocomplete="autocomplete"
      :value="modelValue"
      @input="onInput"
    />
    <button
      type="button"
      class="password-field__toggle"
      :aria-label="visible ? 'Скрыть пароль' : 'Показать пароль'"
      :aria-pressed="visible"
      tabindex="-1"
      @click="toggle"
    >
      <svg
        v-if="!visible"
        class="password-field__icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
        <circle cx="12" cy="12" r="3" />
      </svg>
      <svg
        v-else
        class="password-field__icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M17.94 17.94A10.94 10.94 0 0 1 12 20c-7 0-11-8-11-8a19.77 19.77 0 0 1 5.06-5.94M9.9 4.24A10.94 10.94 0 0 1 12 4c7 0 11 8 11 8a19.87 19.87 0 0 1-2.16 3.19" />
        <path d="M1 1l22 22" />
        <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24" />
      </svg>
    </button>
  </div>
</template>

<script src="./PasswordInput.js"></script>
<style scoped src="./PasswordInput.css"></style>
