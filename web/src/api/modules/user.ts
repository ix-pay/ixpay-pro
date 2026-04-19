import service from '@/utils/request'
import type { ApiResponse, UserInfo, CaptchaResponse } from '@/types'

// @Summary 用户登录
// @Produce  application/json
// @Param data body {userName:"string",password:"string"}
// @Router /auth/login [post]
export const login = (data: {
  userName: string
  password: string
  captcha?: string
  captchaId?: string
}): Promise<ApiResponse<{ user: UserInfo; accessToken?: string; refreshToken?: string }>> => {
  return service({
    url: '/auth/login',
    method: 'post',
    data: data,
  })
}

// @Summary 获取验证码
// @Produce  application/json
// @Router /auth/captcha [post]
export const captcha = (): Promise<ApiResponse<CaptchaResponse>> => {
  return service({
    url: '/auth/captcha',
    method: 'post',
    donNotShowLoading: true, // 不显示全局加载动画
  })
}

// @Summary 用户注册
// @Produce  application/json
// @Param data body {userName:"string",password:"string"}
// @Router /auth/resige [post]
export const register = (data: { userName: string; password: string }): Promise<ApiResponse> => {
  return service({
    url: '/auth/register',
    method: 'post',
    data: data,
  })
}

// @Summary 刷新 token
// @Produce  application/json
// @Param data body {refreshToken:"string"}
// @Router /auth/refresh-token [post]
export const refreshToken = (data: {
  refreshToken: string
}): Promise<ApiResponse<{ accessToken: string; refreshToken?: string }>> => {
  return service({
    url: '/auth/refresh-token',
    method: 'post',
    data: {
      refreshToken: data.refreshToken,
    },
    _isRefreshToken: true, // 标记为刷新 token 的请求
  })
}

// @Summary 修改密码
// @Produce  application/json
// @Param data body {oldPassword:"string",newPassword:"string"}
// @Router /user/password [put]
export const changePassword = (data: {
  oldPassword: string
  newPassword: string
}): Promise<ApiResponse> => {
  return service({
    url: '/user/password',
    method: 'put',
    data: data,
  })
}

// @Tags User
// @Summary 分页获取用户列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param page query number true "页码"
// @Param pageSize query number true "每页条数"
// @Success 200 {string} json "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /user [get]
export const getUserList = (params: {
  page: number
  pageSize: number
  userName?: string
  email?: string
  role?: string
  status?: number
}): Promise<ApiResponse> => {
  return service({
    url: '/user',
    method: 'get',
    params: params,
  })
}

// @Tags User
// @Summary 设置用户权限
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body api.SetUserAuth true "设置用户权限"
// @Success 200 {string} json "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/setUserAuthority [post]
// @Deprecated 此接口后端未实现，已移除
export const setUserAuthority = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口后端未实现')
}

// @Tags SysUser
// @Summary 删除用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "用户 ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/:id [delete]
export const deleteUser = (id: string): Promise<ApiResponse> => {
  return service({
    url: `/user/${id}`,
    method: 'delete',
  })
}

// @Tags SysUser
// @Summary 创建用户
// @Description 创建新用户（管理员权限）
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body {userName:"string",password:"string",email:"string",nickname:"string",phone:"string",avatar:"string",departmentId:"number",positionId:"number",status:"number",roles:"array"} true "创建用户请求参数"
// @Success 201 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /user [post]
export const createUser = (data: {
  userName: string
  password: string
  email: string
  nickname?: string
  phone?: string
  avatar?: string
  departmentId?: number
  positionId?: number
  status?: number
  roles?: string[]
}): Promise<ApiResponse> => {
  return service({
    url: '/user',
    method: 'post',
    data: data,
  })
}

// @Tags SysUser
// @Summary 更新用户信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body {id:"string",nickname:"string",email:"string",phone:"string",avatar:"string",status:"number",roles:"array"} true "更新用户请求参数"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/info [put]
export const updateUserInfo = (data: {
  id: string
  nickname?: string
  email?: string
  phone?: string
  avatar?: string
  status?: number
  roles?: string[]
}): Promise<ApiResponse> => {
  return service({
    url: '/user/info',
    method: 'put',
    data: data,
  })
}

// @Tags SysUser
// @Summary 设置用户信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.SysUser true "设置用户信息"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/setSelfInfo [put]
// @Deprecated 此接口与 updateUserInfo 重复，已移除
export const setSelfInfo = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口与 updateUserInfo 重复')
}

// @Tags SysUser
// @Summary 获取自身界面配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /user/getSelfSetting [get]
// @Deprecated 此接口路径需要修正为 /user/get-user-settings
export const getSelfSetting = (): Promise<ApiResponse> => {
  throw new Error('此接口路径需要修正为 /user/get-user-settings')
}

// @Tags SysUser
// @Summary 设置自身界面配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.SysUser true "设置自身界面配置"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/setSelfSetting [put]
// @Deprecated 此接口路径需要修正为 /user/update-user-settings
export const setSelfSetting = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口路径需要修正为 /user/update-user-settings')
}

// @Tags User
// @Summary 设置用户权限
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body api.setUserAuthorities true "设置用户权限"
// @Success 200 {string} json "{"success":true,"data":{},"msg":"修改成功"}"
// @Router /user/setUserAuthorities [post]
// @Deprecated 此接口后端未实现，已移除
export const setUserAuthorities = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口后端未实现')
}

// @Tags User
// @Summary 获取用户信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} json "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /user/getUserInfo [get]
export const getUserInfo = (): Promise<ApiResponse<{ userInfo: UserInfo }>> => {
  return service({
    url: '/user/info',
    method: 'get',
  })
}

// @Tags SysUser
// @Summary 重置密码
// @Description 重置用户密码（管理员权限）
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body {userId:"string"} true "重置密码请求参数"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"重置成功"}"
// @Router /user/reset-password [put]
export const resetPassword = (data: { userId: string }): Promise<ApiResponse> => {
  return service({
    url: '/user/reset-password',
    method: 'put',
    data: data,
  })
}

// @Summary 用户登出
// @Produce  application/json
// @Router /auth/logout [post]
export const logout = (): Promise<ApiResponse> => {
  return service({
    url: '/auth/logout',
    method: 'post',
  })
}

// @Summary 切换用户角色
// @Description 切换当前用户的活动角色
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body {roleId: string} true "角色 ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"切换成功"}"
// @Router /user/switch-role [post]
export const switchRole = (data: { roleId: string | number }): Promise<ApiResponse> => {
  return service({
    url: '/user/switch-role',
    method: 'post',
    data: data,
  })
}

// 以下为后端未实现或路径需要确认的接口（已废弃）

// @Summary 设置用户权限 (单角色)
// @Description 此接口后端未实现，已废弃
// @Router /user/setUserAuthority [post]
// @Deprecated
export const setUserAuthorityDeprecated = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口后端未实现')
}

// @Summary 设置用户权限 (多角色)
// @Description 此接口后端未实现，已废弃
// @Router /user/setUserAuthorities [post]
// @Deprecated
export const setUserAuthoritiesDeprecated = (
  _data: Record<string, unknown>,
): Promise<ApiResponse> => {
  throw new Error('此接口后端未实现')
}

// @Summary 设置用户信息
// @Description 此接口与 updateUserInfo 重复，已废弃
// @Router /user/setSelfInfo [put]
// @Deprecated
export const setSelfInfoDeprecated = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口与 updateUserInfo 重复')
}

// @Summary 获取自身界面配置
// @Description 此接口路径需要修正为 /user/get-user-settings，已废弃
// @Router /user/getSelfSetting [get]
// @Deprecated
export const getSelfSettingDeprecated = (): Promise<ApiResponse> => {
  throw new Error('此接口路径需要修正为 /user/get-user-settings')
}

// @Summary 设置自身界面配置
// @Description 此接口路径需要修正为 /user/update-user-settings，已废弃
// @Router /user/setSelfSetting [put]
// @Deprecated
export const setSelfSettingDeprecated = (_data: Record<string, unknown>): Promise<ApiResponse> => {
  throw new Error('此接口路径需要修正为 /user/update-user-settings')
}
