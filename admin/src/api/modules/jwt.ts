import service from '@/utils/request'
import type { ApiResponse } from '@/types'

// @Tags jwt
// @Summary jwt 加入黑名单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"拉黑成功"}"
// @Router /auth/jwt/jsonInBlacklist [post]
// 后端路由为 POST /api/admin/auth/jwt/jsonInBlacklist
export const jsonInBlacklist = (): Promise<ApiResponse> => {
  return service({
    url: '/auth/jwt/jsonInBlacklist',
    method: 'post',
  })
}
