import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/main.css'

const app = createApp(App)

// Глобальный обработчик ошибок
app.config.errorHandler = (err, instance, info) => {
  console.error('[Vue Error]', err)
  console.error('[Component]', instance?.$options?.name || 'Unknown')
  console.error('[Info]', info)
}

app.use(createPinia())
app.use(router)
app.mount('#app')
