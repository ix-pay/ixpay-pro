import service from '@/utils/request'
import type { ApiResponse } from '@/types'
import type { ApiMenuItem } from '@/types/menu'

export type MenuItem = ApiMenuItem

// @Tags Menu
// @Summary 获取菜单列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} json "{\"success\":true,\"data\":{},\"msg\":\"获取成功\"}"
// @Router /menu [get]
export const getMenuList = (): Promise<ApiResponse> => {
  return service({
    url: '/menu',
    method: 'get',
  })
}

// @Tags Menu
// @Summary 分页获取菜单列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param page query number true "页码"
// @Param pageSize query number true "每页条数"
// @Success 200 {string} json "{\"success\":true,\"data\":{},\"msg\":\"获取成功\"}"
// @Router /menu/page [get]
export const getMenuPage = (params: { page: number; pageSize: number }): Promise<ApiResponse> => {
  return service({
    url: '/menu/page',
    method: 'get',
    params: params,
  })
}

// @Tags Menu
// @Summary 添加菜单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body MenuItem true "添加菜单"
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"添加成功\"}"
// @Router /menu [post]
export const addMenu = (data: MenuItem): Promise<ApiResponse> => {
  return service({
    url: '/menu',
    method: 'post',
    data: data,
  })
}

// @Tags Menu
// @Summary 更新菜单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body MenuItem true "更新菜单"
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"修改成功\"}"
// @Router /menu [put]
export const updateMenu = (data: MenuItem): Promise<ApiResponse> => {
  return service({
    url: '/menu',
    method: 'put',
    data: data,
  })
}

// @Tags Menu
// @Summary 删除菜单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "菜单 ID"
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"删除成功\"}"
// @Router /menu/:id [delete]
export const deleteMenu = (id: string): Promise<ApiResponse> => {
  return service({
    url: `//menu/${id}`,
    method: 'delete',
  })
}

// @Tags Menu
// @Summary 获取菜单树结构
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"获取成功\"}"
// @Router /menu/tree [get]
export const getMenuTree = (): Promise<ApiResponse> => {
  return service({
    url: '/menu/tree',
    method: 'get',
  })
}

// @Tags API
// @Summary 获取所有 API 列表（用于菜单关联 API 选择）
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"获取成功\"}"
// @Router /apis [get]
export const getApiList = (): Promise<ApiResponse> => {
  return service({
    url: '/apis',
    method: 'get',
  })
}

// @Tags API
// @Summary 搜索 API 列表（支持分页和关键词搜索）
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param page query number false "页码"
// @Param pageSize query number false "每页条数"
// @Param keyword query string false "搜索关键词"
// @Success 200 {string} string "{\"success\":true,\"data\":{},\"msg\":\"获取成功\"}"
// @Router /apis [get]
export const searchApiList = (params?: {
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<ApiResponse> => {
  // 搜索时默认获取所有匹配结果，避免分页导致搜索结果不完整
  const defaultParams = {
    page: 1,
    pageSize: 1000, // 搜索时使用较大的 pageSize，确保获取所有匹配项
    ...params,
  }

  return service({
    url: '/apis',
    method: 'get',
    params: defaultParams,
  })
}
