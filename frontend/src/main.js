import {createApp} from 'vue'
import App from './App.vue'
import './styles/global.css' // Импорт глобальных стилей

const app = createApp(App)

app.directive('focus', {
    mounted: (el) => el.focus(),
})

app.mount('#app')