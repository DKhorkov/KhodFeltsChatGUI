export default {
    name: 'ProfileModal',
    props: {
        user: {
            type: Object,
            required: true,
        },
        isDarkTheme: {
            type: Boolean,
            required: true,
        },
    },
    emits: ['close', 'toggle-theme', 'logout'],

    setup() {
        const formatDate = (dateStr) => {
            return new Date(dateStr).toLocaleDateString('ru-RU', {
                day: 'numeric',
                month: 'long',
                year: 'numeric',
            })
        }

        return {
            formatDate,
        }
    },
}
