<template>
  <div id="app">
    <el-config-provider :locale="locale">
      <router-view />
      <Application />
    </el-config-provider>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import Application from '@/components/business/application/index.vue'

defineOptions({
  name: 'App',
})

// 多语言切换功能
const language = ref('zh-cn')
const locale = computed(() => (language.value === 'zh-cn' ? zhCn : en))
</script>

<style scoped lang="scss">
#app {
  height: 100vh;
  overflow: hidden;
}
</style>
