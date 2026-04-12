/**
 * 业务组件库导出
 *
 * 基于 Element Plus + Tailwind CSS + 设计系统的混合架构
 * 提供可复用的高级业务组件
 */

// 卡片组件
export { default as IxCard } from './Card/index.vue'

// 统计卡片组件
export { default as IxStatCard } from './StatCard/index.vue'

// 页面模板组件
export { default as IxPageTemplate } from './PageTemplate/index.vue'

// 主题配置面板
export { default as ThemePanel } from './ThemePanel/index.vue'

// 图表组件
export { default as IxEcharts } from './Chart/IxEcharts.vue'

// 导入所有组件（用于 Vue.use）
import IxCard from './Card/index.vue'
import IxStatCard from './StatCard/index.vue'
import IxPageTemplate from './PageTemplate/index.vue'
import ThemePanel from './ThemePanel/index.vue'
import IxEcharts from './Chart/IxEcharts.vue'

export const IxBusinessComponents = [IxCard, IxStatCard, IxPageTemplate, ThemePanel, IxEcharts]

// 默认导出
export default {
  install(app: unknown) {
    IxBusinessComponents.forEach((component) => {
      if (app && typeof app === 'object' && 'component' in app) {
        ;(app as { component: (name: string, component: unknown) => void }).component(
          component.name || '',
          component,
        )
      }
    })
  },
}
