import {ref} from 'vue'

export default {
    name: 'PasswordInput',

    props: {
        modelValue: {type: String, default: ''},
        placeholder: {type: String, default: ''},
        id: {type: String, default: ''},
        name: {type: String, default: ''},
        required: {type: Boolean, default: false},
        autocomplete: {type: String, default: 'current-password'},
        inputClass: {type: String, default: ''},
    },

    emits: ['update:modelValue'],

    setup(props, {emit}) {
        const visible = ref(false)

        const toggle = () => {
            visible.value = !visible.value
        }

        const onInput = (event) => {
            emit('update:modelValue', event.target.value)
        }

        return {visible, toggle, onInput}
    },
}
