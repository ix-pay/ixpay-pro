<template>
  <div class="setting-page">
    <h2>系统设置</h2>
    <div class="setting-content">
      <el-card shadow="hover">
        <template #header>
          <div class="card-header">
            <span>界面设置</span>
          </div>
        </template>
        <el-form
          ref="settingFormRef"
          :model="settingForm"
          :rules="formRules"
          label-width="120px"
          class="setting-form"
        >
          <el-form-item label="主题模式" prop="darkMode">
            <el-radio-group v-model="settingForm.darkMode">
              <el-radio-button label="light">浅色</el-radio-button>
              <el-radio-button label="dark">深色</el-radio-button>
              <el-radio-button label="auto">自动</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="主题颜色" prop="primaryColor">
            <el-color-picker
              v-model="settingForm.primaryColor"
              show-color-palette
              :predefine="themeColorOptions"
            />
          </el-form-item>
          <el-form-item label="字体大小" prop="fontSize">
            <el-slider v-model="settingForm.fontSize" :min="12" :max="20" :step="1" />
          </el-form-item>
          <el-form-item label="侧边栏宽度" prop="layout_side_width">
            <el-slider
              v-model="settingForm.layout_side_width"
              :min="180"
              :max="300"
              :step="10"
              :marks="{
                180: '窄',
                240: '标准',
                300: '宽',
              }"
            />
          </el-form-item>
          <el-form-item label="显示标签栏">
            <el-switch v-model="settingForm.showTabs" />
          </el-form-item>
          <el-form-item label="显示水印">
            <el-switch v-model="settingForm.show_watermark" />
          </el-form-item>
          <el-form-item label="语言" prop="language">
            <el-select v-model="settingForm.language" placeholder="请选择语言">
              <el-option label="简体中文" value="zh-CN" />
              <el-option label="English" value="en-US" />
            </el-select>
          </el-form-item>
          <el-form-item label="自动登录">
            <el-switch v-model="settingForm.autoLogin" />
          </el-form-item>
          <el-form-item label="记住密码">
            <el-switch v-model="settingForm.rememberPassword" />
          </el-form-item>
        </el-form>
        <div class="form-actions">
          <el-button type="primary" @click="handleSaveSettings">保存设置</el-button>
          <el-button @click="handleResetSettings">重置</el-button>
        </div>
      </el-card>

      <!-- 系统信息卡片 -->
      <el-card shadow="hover" style="margin-top: 20px">
        <template #header>
          <div class="card-header">
            <span>系统信息</span>
          </div>
        </template>
        <div class="system-info">
          <div class="info-item">
            <label>系统版本</label>
            <span>{{ systemInfo.version }}</span>
          </div>
          <div class="info-item">
            <label>前端框架</label>
            <span>Vue 3 + TypeScript</span>
          </div>
          <div class="info-item">
            <label>后端框架</label>
            <span>Go</span>
          </div>
          <div class="info-item">
            <label>最后更新</label>
            <span>{{ systemInfo.lastUpdate }}</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '@/stores'
import { storeToRefs } from 'pinia'

import { getSelfSetting, setSelfSetting } from '@/api/modules/user'

defineOptions({
  name: 'SettingPage',
})

interface SettingForm {
  darkMode: string
  primaryColor: string
  fontSize: number
  layout_side_width: number
  showTabs: boolean
  show_watermark: boolean
  language: string
  autoLogin: boolean
  rememberPassword: boolean
}

interface SystemInfo {
  version: string
  lastUpdate: string
}

// 使用 appStore
const appStore = useAppStore()
const { config } = storeToRefs(appStore)

// 加载状态
const loading = ref(false)
// 表单引用
const settingFormRef = ref()

// 主题颜色选项
const themeColorOptions = [
  '#3b82f6', // 蓝色（默认）
  '#67C23A', // 绿色
  '#E6A23C', // 橙色
  '#F56C6C', // 红色
  '#909399', // 灰色
  '#722ED1', // 紫色
  '#13C2C2', // 青色
  '#FAAD14', // 黄色
]

// 设置表单数据
const settingForm = reactive<SettingForm>({
  darkMode: 'auto',
  primaryColor: '#3b82f6',
  fontSize: 14,
  layout_side_width: 256,
  showTabs: true,
  show_watermark: true,
  language: 'zh-CN',
  autoLogin: false,
  rememberPassword: true,
})

// 系统信息
const systemInfo = reactive<SystemInfo>({
  version: '1.0.0',
  lastUpdate: '2024-01-01',
})

// 表单验证规则
const formRules = reactive({
  darkMode: [{ required: true, message: '请选择主题模式', trigger: 'change' }],
  primaryColor: [{ required: true, message: '请选择主题颜色', trigger: 'change' }],
  fontSize: [{ required: true, message: '请设置字体大小', trigger: 'change' }],
  layout_side_width: [{ required: true, message: '请设置侧边栏宽度', trigger: 'change' }],
  language: [{ required: true, message: '请选择语言', trigger: 'change' }],
})

// 获取设置信息
const loadSettings = async () => {
  loading.value = true
  try {
    // 从 appStore 获取配置
    settingForm.darkMode = config.value.darkMode
    settingForm.primaryColor = config.value.primaryColor
    settingForm.fontSize = 14 // 字体大小不从 store 获取，使用默认值
    settingForm.layout_side_width = config.value.layout_side_width
    settingForm.showTabs = config.value.showTabs
    settingForm.show_watermark = config.value.show_watermark

    // 从服务器获取用户设置
    try {
      const res = await getSelfSetting()
      if ((res.data as Record<string, unknown>)?.settings) {
        const userSettings = (res.data as Record<string, unknown>).settings as SettingForm
        settingForm.language = userSettings.language || 'zh-CN'
        settingForm.autoLogin = userSettings.autoLogin || false
        settingForm.rememberPassword =
          userSettings.rememberPassword !== undefined ? userSettings.rememberPassword : true
      }
    } catch (apiError) {
      console.error('从服务器获取设置失败:', apiError)
    }
  } catch (error) {
    console.error('获取设置信息失败:', error)
  } finally {
    loading.value = false
  }
}

// 保存设置
const handleSaveSettings = async () => {
  try {
    await settingFormRef.value.validate()

    // 使用 appStore 的方法更新配置
    if (settingForm.darkMode === 'light') {
      appStore.toggleLightMode()
    } else if (settingForm.darkMode === 'dark') {
      appStore.toggleDarkModeForce()
    } else if (settingForm.darkMode === 'auto') {
      appStore.toggleAutoTheme()
    }

    appStore.togglePrimaryColor(settingForm.primaryColor)
    appStore.toggleConfigSideWidth(settingForm.layout_side_width)
    appStore.toggleTabs(settingForm.showTabs)
    appStore.toggleConfigWatermark(settingForm.show_watermark)

    // 保存用户设置到 localStorage
    const userSettings = {
      language: settingForm.language,
      autoLogin: settingForm.autoLogin,
      rememberPassword: settingForm.rememberPassword,
    }
    localStorage.setItem('userSettings', JSON.stringify(userSettings))

    // 调用 API 保存到服务器
    try {
      await setSelfSetting(userSettings)
    } catch (error) {
      console.error('保存设置到服务器失败:', error)
    }

    ElMessage.success('设置保存成功')
  } catch (error) {
    console.error('保存设置失败:', error)
    ElMessage.error('保存设置失败')
  }
}

// 重置设置
const handleResetSettings = () => {
  // 重置为默认值
  Object.assign(settingForm, {
    darkMode: 'auto',
    primaryColor: '#3b82f6',
    fontSize: 14,
    layout_side_width: 256,
    showTabs: true,
    show_watermark: true,
    language: 'zh-CN',
    autoLogin: false,
    rememberPassword: true,
  })
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.setting-page {
  background-color: var(--bg-color);
  padding: 20px;
  min-height: calc(100vh - 60px);
  color: var(--text-primary);
  transition:
    background-color 0.3s ease,
    color 0.3s ease;
}

.setting-content {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  font-weight: bold;
  font-size: 16px;
  color: var(--text-primary);
}

.setting-form {
  margin-top: 20px;
}

.form-actions {
  margin-top: 30px;
  text-align: right;
}

.system-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 5px 0;
  color: var(--text-primary);
}

.info-item label {
  font-weight: bold;
  color: var(--text-primary);
}
</style>
