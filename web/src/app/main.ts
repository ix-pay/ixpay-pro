import { createApp } from 'vue'
import ElementPlus from 'element-plus'

import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
// 引入 Tailwind CSS
import '@/assets/styles/tailwind.css'
// 引入设计令牌系统（必须在 global.scss 之前）
import '@/assets/styles/design-tokens.scss'
// 引入全局样式
import '@/assets/styles/global.scss'

import run from '@/core/ixpay-pro'
import auth from '@/directive/auth'
import authBtn from '@/directive/auth-btn'
import { store } from '@/stores'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(run).use(ElementPlus).use(store).use(router).use(auth).use(authBtn).mount('#app')
