// 任务调度相关类型定义
// 所有 ID 字段使用 string 类型，避免 JavaScript Number 精度丢失问题

// 任务信息
export interface Task {
  id: string // 任务 ID
  taskId: string // 任务标识
  taskType: string // 任务类型（http/database/cache/script）
  type: string // 调度类型（cron/one_time）
  expression: string // Cron 表达式或执行时间
  description: string // 任务描述
  group: string // 任务分组
  status: string // 数据库状态（enabled/disabled）
  statusLabel: string // 运行状态（running/stopped）
  createdAt: string // 创建时间
  lastRunAt: string // 最后执行时间
  nextRunAt: string // 下次执行时间
  retryCount: number // 重试次数
  maxRetries: number // 最大重试次数
  concurrency?: string // 并发模式（allow/skip/wait）
  timeout?: number // 超时时间（秒）
  assignedNode?: string // 分配节点
  nodeGroup?: string // 节点组
}

// 任务日志
export interface TaskLog {
  id: string // 日志 ID
  taskId: string // 任务 ID
  taskName: string // 任务名称
  group: string // 任务分组
  executeAt: string // 执行时间
  duration: number // 执行耗时（毫秒）
  result: string // 执行结果（success/failed）
  errorInfo: string // 错误信息
  retryCount: number // 重试次数
  cronExpr: string // Cron 表达式
  triggerType: string // 触发类型
  operatorId: string // 操作人 ID
}

// 任务统计
export interface TaskStatistics {
  taskId: string // 任务 ID
  taskName: string // 任务名称
  group: string // 任务分组
  totalExecutes: number // 总执行次数
  successCount: number // 成功次数
  failedCount: number // 失败次数
  successRate: number // 成功率
  avgDuration: number // 平均耗时（毫秒）
  lastExecuteAt: string // 最后执行时间
  nextExecuteAt: string // 下次执行时间
}

// 任务统计面板
export interface TaskDashboard {
  totalTasks: number // 任务总数
  enabledTasks: number // 启用数
  disabledTasks: number // 禁用数
  todayExecutions: number // 今日执行数
}

// 创建任务请求
export interface CreateTaskRequest {
  taskId: string // 任务标识
  taskType: string // 任务类型
  type: string // 调度类型
  expression: string // 表达式
  description: string // 任务描述
  group?: string // 任务分组
  retryCount: number // 重试次数
  params: Record<string, unknown> // 任务参数
}

// 更新任务请求
export interface UpdateTaskRequest {
  taskId: string // 任务标识
  taskType: string // 任务类型
  type: string // 调度类型
  expression: string // 表达式
  description: string // 任务描述
  group?: string // 任务分组
  retryCount: number // 重试次数
  params: Record<string, unknown> // 任务参数
}

// 任务搜索参数
export interface TaskSearchParams {
  page?: number // 页码
  pageSize?: number // 每页条数
  taskId?: string // 任务 ID
  taskType?: string // 任务类型
  group?: string // 任务分组
  status?: string // 状态
}
