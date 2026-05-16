import service from '@/utils/request'
import type { ApiResponse } from '@/types'

// @Tags jwt
// @Summary jwt 加入黑名单
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"拉黑成功"}"
// @Router /jwt/jsonInBlacklist [post]
export const jsonInBlacklist = (): Promise<ApiResponse> => {
  return service({
    url: '/jwt/jsonInBlacklist',
    method: 'post',
  })
}
