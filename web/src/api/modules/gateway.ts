import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface GatewayService {
  id: string
  name: string
  address: string
  port: number
  metadata: Record<string, string>
  lastSeen: string
  activeConnections: number
  status: string
}

export const getGatewayServices = (): Promise<ApiResponse<GatewayService[]>> => {
  return service({
    url: '/gateway/services',
    method: 'get'
  })
}
