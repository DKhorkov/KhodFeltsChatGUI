import {onUnmounted} from 'vue'

export default {
    name: 'NotificationToast', props: {
        message: {
            type: String, required: true,
        },
        duration: {
            type: Number, default: 3000,
        },
    }, emits: ['close', 'click'],

    setup(props, {emit}) {
        const timer = setTimeout(() => emit('close'), props.duration)
        onUnmounted(() => clearTimeout(timer))
    },
}
