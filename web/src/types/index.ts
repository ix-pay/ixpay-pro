// 通用类型定义

// API 响应通用结构
export interface ApiResponse<T = unknown> {
  code: number
  data?: T
  msg?: string
}

// 分页响应结构
export interface PageResponse<T = unknown> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

// 角色信息接口（简化版）
export interface RoleInfo {
  id: string // 角色 ID，使用 string 避免精度丢失
  name: string // 角色名称
  code: string // 角色编码
  description?: string // 描述
  type?: number // 角色类型
  parentId?: string // 父角色 ID
  status?: number // 状态
  isSystem?: boolean // 是否系统角色
  sort?: number // 排序
}

// 用户信息接口（简化版，用于通用场景）
export interface UserInfo {
  id: string // 用户 ID
  userName: string // 用户名
  nickname: string // 昵称
  email?: string // 邮箱
  phone?: string // 手机号
  avatar?: string // 头像
  status?: number // 状态
  roles?: RoleInfo[] // 角色列表
  createdAt?: string // 创建时间
  updatedAt?: string // 更新时间
  // 兼容后端返回的字段和额外字段
  authorityId?: number | string // 权限 ID
  currentRoleId?: string | number // 当前角色 ID
  role?: string // 角色名称（兼容旧版）
  authority?: {
    permissions?: string[]
    roles?: string[]
    defaultRouter?: string // 默认路由
  }
  uuid?: string
  nickName?: string // 兼容旧版字段
  headerImg?: string
}

// 登录请求参数
export interface LoginRequest {
  userName: string
  password: string
  captcha?: string
  captchaId?: string
}

// 登录响应
export interface LoginResponse {
  user: UserInfo
  accessToken: string
  refreshToken: string
}

// 验证码响应
export interface CaptchaResponse {
  captchaId: string
  picPath: string
  captchaLength: number
  openCaptcha: boolean
}

// 刷新 Token 请求
export interface RefreshTokenRequest {
  refreshToken: string
}

// 菜单信息接口（简化版）
export interface MenuItem {
  id: string
  parentId: string
  name: string
  path: string
  component?: string
  icon?: string
  sort?: number
  status?: number
  type?: number
  children?: MenuItem[]
}

// 字典信息接口（简化版）
export interface DictItem {
  id: string
  dictId: string
  itemKey: string
  itemValue: string
  sort?: number
  status?: number
}

// 配置信息接口（简化版）
export interface ConfigItem {
  id: string
  configKey: string
  configValue: string
  configType: string
  status?: number
}

// 部门信息接口（简化版）
export interface DepartmentItem {
  id: string
  parentId: string
  name: string
  sort?: number
  status?: number
  children?: DepartmentItem[]
}

// 岗位信息接口（简化版）
export interface PositionItem {
  id: string
  name: string
  code: string
  sort?: number
  status?: number
}

// 扩展Vue的类型声明，添加$IXPAY_PRO的类型
declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $IXPAY_PRO: {
      appName: string
      appLogo: string
      showViteLogo: boolean
      logs: unknown[]
    }
  }
}
