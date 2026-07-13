import {createApp} from 'vue'
import App from './App.vue'
import router from './router/router'
import './styles/tokens.css'
import './styles/base.css'
import { applyThemeMode, THEME_MODE_STORAGE_KEY } from './utils/theme-mode.mjs'
// 引入组件库的少量全局样式变量
import 'tdesign-vue-next/es/style/index.css';

const storedThemeMode = window.localStorage.getItem(THEME_MODE_STORAGE_KEY)
applyThemeMode(storedThemeMode)

const app = createApp(App)

app.config.errorHandler = (err) => {
  if (err.message && err.message.includes('ResizeObserver')) {
    return
  }
  console.error(err)
}

window.addEventListener('error', (event) => {
  if (event.message && event.message.includes('ResizeObserver')) {
    event.preventDefault()
    return true
  }
})

window.addEventListener('unhandledrejection', (event) => {
  if (event.reason && event.reason.message && event.reason.message.includes('ResizeObserver')) {
    event.preventDefault()
    return true
  }
})

app.use(router)
app.mount('#app')
