import 'element-plus/theme-chalk/dark/css-vars.css'

import { createApp } from 'vue'
import ElementPlus from 'element-plus'

import 'element-plus/dist/index.css'

import run from '@/core/ixpay-pro.js'
import auth from '@/directive/auth'
import { store } from '@/pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(run).use(ElementPlus).use(store).use(auth).use(router).mount('#app')
