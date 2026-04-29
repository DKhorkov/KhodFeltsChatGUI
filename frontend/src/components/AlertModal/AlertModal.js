export default {
    name: 'AlertModal', props: {
        message: {
            type: String, required: true,
        },
        type: {
            type: String, default: 'info',
            validator: (v) => ['error', 'info'].includes(v),
        },
    }, emits: ['close'],
}
