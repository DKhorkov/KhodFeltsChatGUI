export default {
  name: 'NotificationToast',
  props: {
    message: {
      type: String,
      required: true
    },
    duration: {
      type: Number,
      default: 3000
    }
  },
  emits: ['close'],

  setup(props, { emit }) {
    // Автоматическое закрытие через указанное время
    const timer = setTimeout(() => {
      emit('close')
    }, props.duration)

    // Очищаем таймер при размонтировании
    const cleanup = () => {
      clearTimeout(timer)
    }

    return {
      cleanup
    }
  },

  beforeUnmount() {
    this.cleanup()
  }
}