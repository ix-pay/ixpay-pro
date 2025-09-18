import config from './config'

// 统一导入el-icon图标
import * as ElementPlusIconsVue from '@element-plus/icons-vue'


export const register = (app) => {
  // 统一注册el-icon图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }
  app.config.globalProperties.$IXPAY_PRO = config
}
