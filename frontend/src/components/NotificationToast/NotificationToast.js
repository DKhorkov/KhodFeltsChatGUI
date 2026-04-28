export default {
    name: 'NotificationToast', props: {
        message: {
            type: String, required: true,
        },
    }, emits: ['close', 'click'],
}
