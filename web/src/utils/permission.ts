import { useUserStore } from '@/stores/modules/user'
import { useRouterStore } from '@/stores/modules/router'
import type { ApiMenuItem } from '@/types/menu'

/**
 * 检查用户是否有权限
 * @param permission 权限标识
 * @returns 是否有权限
 */
export const hasPermission = (permission: string): boolean => {
  const userStore = useUserStore()
  const userInfo = userStore.userInfo

  // 获取用户的权限列表
  const permissions = (userInfo?.authority?.permissions as string[]) || []

  // 如果用户没有权限列表，默认没有权限
  if (!permissions || !Array.isArray(permissions) || permissions.length === 0) {
    return false
  }

  // 检查用户是否有对应的权限
  return permissions.includes(permission)
}

/**
 * 检查用户是否有多个权限中的任意一个
 * @param permissions 权限标识数组
 * @returns 是否有权限
 */
export const hasAnyPermission = (permissions: string[]): boolean => {
  if (!permissions || !Array.isArray(permissions)) {
    return false
  }

  return permissions.some((permission) => hasPermission(permission))
}

/**
 * 检查用户是否有所有权限
 * @param permissions 权限标识数组
 * @returns 是否有权限
 */
export const hasAllPermissions = (permissions: string[]): boolean => {
  if (!permissions || !Array.isArray(permissions)) {
    return false
  }

  return permissions.every((permission) => hasPermission(permission))
}

/**
 * 检查用户是否有菜单权限
 * @param menu 菜单对象
 * @returns 是否有权限
 */
export const hasMenuPermission = (menu: ApiMenuItem): boolean => {
  // 如果菜单没有 permission 字段，默认有权限
  if (!menu || !menu.permission) {
    return true
  }

  return hasPermission(menu.permission)
}

/**
 * 获取用户的权限列表
 * @returns 权限列表
 */
export const getPermissions = (): string[] => {
  const userStore = useUserStore()
  const userInfo = userStore.userInfo

  return (userInfo?.authority?.permissions as string[]) || []
}

/**
 * 检查用户是否有角色
 * @param role 角色标识
 * @returns 是否有角色
 */
export const hasRole = (role: string): boolean => {
  const userStore = useUserStore()
  const userInfo = userStore.userInfo

  // 获取用户的角色列表
  const roles = (userInfo?.authority?.roles as string[]) || []

  // 如果用户没有角色列表，默认没有角色
  if (!roles || !Array.isArray(roles) || roles.length === 0) {
    return false
  }

  // 检查用户是否有对应的角色
  return roles.includes(role)
}

/**
 * 检查用户是否有多个角色中的任意一个
 * @param roles 角色标识数组
 * @returns 是否有角色
 */
export const hasAnyRole = (roles: string[]): boolean => {
  if (!roles || !Array.isArray(roles)) {
    return false
  }

  return roles.some((role) => hasRole(role))
}

/**
 * 检查用户是否有所有角色
 * @param roles 角色标识数组
 * @returns 是否有角色
 */
export const hasAllRoles = (roles: string[]): boolean => {
  if (!roles || !Array.isArray(roles)) {
    return false
  }

  return roles.every((role) => hasRole(role))
}

/**
 * 获取用户的角色列表
 * @returns 角色列表
 */
export const getRoles = (): string[] => {
  const userStore = useUserStore()
  const userInfo = userStore.userInfo

  return (userInfo?.authority?.roles as string[]) || []
}

/**
 * 检查用户是否有按钮权限
 * @param permission 按钮权限标识
 * @returns 是否有权限
 */
export const hasButtonPermission = (permission: string): boolean => {
  const routerStore = useRouterStore()

  // 从路由 store 中获取按钮权限列表（从菜单数据中提取的 type: 3 的按钮权限）
  const buttonPermissions = routerStore.buttonPermissions

  // 如果没有按钮权限列表，默认没有权限
  if (!buttonPermissions || !Array.isArray(buttonPermissions) || buttonPermissions.length === 0) {
    return false
  }

  // 检查用户是否有对应的按钮权限
  return buttonPermissions.includes(permission)
}

/**
 * 检查用户是否有多个按钮权限中的任意一个
 * @param permissions 权限标识数组
 * @returns 是否有权限
 */
export const hasAnyButtonPermission = (permissions: string[]): boolean => {
  if (!permissions || !Array.isArray(permissions)) {
    return false
  }

  return permissions.some((permission) => hasButtonPermission(permission))
}

/**
 * 检查用户是否有所有按钮权限
 * @param permissions 权限标识数组
 * @returns 是否有权限
 */
export const hasAllButtonPermissions = (permissions: string[]): boolean => {
  if (!permissions || !Array.isArray(permissions)) {
    return false
  }

  return permissions.every((permission) => hasButtonPermission(permission))
}

/**
 * 获取用户的按钮权限列表
 * @returns 按钮权限列表
 */
export const getButtonPermissions = (): string[] => {
  const routerStore = useRouterStore()
  return routerStore.buttonPermissions
}
