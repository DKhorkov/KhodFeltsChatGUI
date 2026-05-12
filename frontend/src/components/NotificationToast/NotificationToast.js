export default {
    name: 'NotificationToast', props: {
        message: {
            type: String, required: true,
        },
        sender: {
            type: String, default: '',
        },
    }, emits: ['close', 'click'],
}
