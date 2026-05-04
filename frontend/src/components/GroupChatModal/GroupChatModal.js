import {computed} from 'vue'

export default {
    name: 'GroupChatModal',
    props: {
        chat: {
            type: Object,
            required: true,
        },
        currentUser: {
            type: Object,
            required: true,
        },
    },
    emits: ['close', 'open-member-profile'],

    setup(props, {emit}) {
        const chatTitle = computed(() => {
            return props.chat.title || `Чат #${props.chat.id}`
        })

        const chatInitial = computed(() => {
            return chatTitle.value.charAt(0).toUpperCase()
        })

        const membersCount = computed(() => {
            return props.chat.members ? props.chat.members.length : 0
        })

        const openMemberProfile = (member) => {
            emit('open-member-profile', member)
        }

        return {
            chatTitle,
            chatInitial,
            membersCount,
            openMemberProfile,
        }
    },
}
