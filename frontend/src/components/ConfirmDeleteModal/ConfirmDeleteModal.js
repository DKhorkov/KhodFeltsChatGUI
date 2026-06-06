export default {
    name: 'ConfirmDeleteModal',
    props: {
        message: {
            type: String,
            required: true,
        },
        title: {
            type: String,
            default: 'Подтверждение',
        },
        confirmText: {
            type: String,
            default: 'Удалить',
        },
        cancelText: {
            type: String,
            default: 'Отмена',
        },
        confirmType: {
            type: String,
            default: 'danger',
            validator: (v) => ['danger', 'primary'].includes(v),
        },
    },
    emits: ['confirm', 'cancel'],
}
