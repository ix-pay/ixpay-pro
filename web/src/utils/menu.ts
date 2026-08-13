/**
 * 菜单路径工具函数
 * 统一处理左侧和顶部菜单的路径构建，确保格式一致
 */

// 获取菜单项完整路径（拼接父级和子级路径）
export const getFullPath = (parentPath: string, childPath: string): string => {
  const cleanParentPath = parentPath.endsWith('/') ? parentPath.slice(0, -1) : parentPath
  const cleanChildPath = childPath.startsWith('/') ? childPath.slice(1) : childPath
  return `/${cleanParentPath}/${cleanChildPath}`
}

// 获取正确的菜单路径（确保以 / 开头）
export const getMenuPath = (path: string): string => {
  if (!path) return '/'
  if (path.startsWith('/')) return path
  return `/${path}`
}