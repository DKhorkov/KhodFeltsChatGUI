import {computed, ref} from 'vue'

const CATEGORIES = [
    {
        name: 'smileys',
        label: 'Смайлы',
        icon: '😀',
        emojis: [
            '😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂',
            '🙂', '😉', '😊', '😇', '🥰', '😍', '🤩', '😘',
            '😋', '😛', '😜', '🤪', '😝', '🤗', '🤔', '🤭',
            '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '😮‍💨',
            '🤥', '😌', '😔', '😪', '🤤', '😴', '😷', '🤒',
            '🤕', '🥴', '😵', '🤯', '🥳', '🥺', '😢', '😭',
            '😤', '😠', '😡', '🤬', '😈', '👿', '💀', '☠️',
        ],
    },
    {
        name: 'gestures',
        label: 'Жесты',
        icon: '👋',
        emojis: [
            '👋', '🤚', '🖐️', '✋', '🖖', '👌', '🤌', '🤏',
            '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆',
            '👇', '☝️', '👍', '👎', '✊', '👊', '🤛', '🤜',
            '👏', '🙌', '👐', '🤲', '🤝', '🙏', '💪', '🦾',
        ],
    },
    {
        name: 'hearts',
        label: 'Сердца',
        icon: '❤️',
        emojis: [
            '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍',
            '🤎', '💔', '❤️‍🔥', '💕', '💞', '💓', '💗', '💖',
            '💘', '💝', '💟', '♥️', '❣️', '💋', '🫶', '💑',
        ],
    },
    {
        name: 'objects',
        label: 'Объекты',
        icon: '🎉',
        emojis: [
            '🎉', '🎊', '🎈', '🎁', '🏆', '🥇', '🎯', '🔥',
            '⭐', '🌟', '✨', '💫', '🌈', '☀️', '🌙', '⚡',
            '💡', '📌', '📎', '✏️', '📝', '💬', '💭', '🗨️',
            '✅', '❌', '⚠️', '❓', '❗', '💯', '🆗', '🔔',
        ],
    },
]

export default {
    name: 'EmojiPicker',
    emits: ['select'],

    setup() {
        const categories = CATEGORIES
        const activeCategory = ref(CATEGORIES[0].name)

        const activeEmojis = computed(() => {
            return categories.find(c => c.name === activeCategory.value)?.emojis ?? []
        })

        return {
            categories,
            activeCategory,
            activeEmojis,
        }
    },
}
