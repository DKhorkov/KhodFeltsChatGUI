import {createApp} from 'vue'
import App from './App.vue'
import './styles/global.css' // Импорт глобальных стилей

const app = createApp(App)

// v-focus: переводит фокус на элемент при монтировании и возвращает на
// предыдущий focused-элемент при размонтировании. Это нужно, чтобы при закрытии
// модалки/оверлея фокус возвращался на родителя (другую модалку, область чата
// и т. п.), а не уходил на body — иначе Escape перестаёт работать.
//
// Фолбэк: если activeElement был body или null (например, когда перед открытием
// модалки умерла фокусированная <button> из dropdown'а), ищем ближайший
// родительский .modal-overlay — это вернёт фокус на «обёртывающую» модалку.
app.directive('focus', {
    mounted: (el) => {
        const prev = document.activeElement
        if (prev && prev !== document.body && prev !== el && document.contains(prev)) {
            el._previousFocus = prev
        } else {
            el._previousFocus = el.parentElement
                ? el.parentElement.closest('.modal-overlay[tabindex]')
                : null
        }
        el.focus()
    },
    unmounted: (el) => {
        const prev = el._previousFocus
        if (prev && typeof prev.focus === 'function' && document.contains(prev)) {
            prev.focus()
        }
        el._previousFocus = null
    },
})

app.mount('#app')