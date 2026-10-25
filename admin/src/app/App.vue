<template>
  <div id="app">
    <el-config-provider :locale="locale">
      <router-view />
    </el-config-provider>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores'
import { store } from '@/stores'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

defineOptions({
  name: 'App',
})

const appStore = useAppStore(store)
const { config } = storeToRefs(appStore)

// 多语言切换，从 appStore 配置读取
const locale = computed(() => (config.value.language === 'en-US' ? en : zhCn))
</script>

<style scoped lang="scss">
#app {
  height: 100vh;
  overflow: hidden;
}
</style>
