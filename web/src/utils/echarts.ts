import * as echarts from 'echarts/core'
import {
  BarChart,
  LineChart,
  PieChart,
  RadarChart,
  MapChart,
  PictorialBarChart,
  GaugeChart,
} from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  PolarComponent,
  AriaComponent,
  ParallelComponent,
  LegendComponent,
  LegendScrollComponent,
  DataZoomComponent,
  ToolboxComponent,
  VisualMapComponent,
  TransformComponent,
} from 'echarts/components'
import { CanvasRenderer, SVGRenderer } from 'echarts/renderers'

// 注册必需的组件
echarts.use([
  // 图表类型
  BarChart,
  LineChart,
  PieChart,
  RadarChart,
  MapChart,
  PictorialBarChart,
  GaugeChart,
  // 组件
  TitleComponent,
  TooltipComponent,
  GridComponent,
  PolarComponent,
  AriaComponent,
  ParallelComponent,
  LegendComponent,
  LegendScrollComponent,
  DataZoomComponent,
  ToolboxComponent,
  VisualMapComponent,
  TransformComponent,
  // 渲染器
  CanvasRenderer,
  SVGRenderer,
])

// 导出 echarts 实例
export { echarts }
export type { EChartsOption } from 'echarts'
export type { ECharts } from 'echarts'
