import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import 'vue-toast-notification/dist/theme-sugar.css'

createApp(App).use(router).mount('#app')