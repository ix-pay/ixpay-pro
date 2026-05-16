import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Dict {
  id: number
  dictName: string
  dictCode: string
  description: string
  status: number
  createdAt: string
  updatedAt: string
  itemCount?: number
  dictItems?: DictItem[]
}

export interface DictItem {
  id: number
  dictId: number
  itemKey: string
  itemValue: string
  sort: number
  description: string
  status: number
  createdAt: string
  updatedAt: string
}

/**
 * 获取字典列表（分页）
 * 使用场景：字典管理页面主列表
 */
export const getDictList = (params?: {
  page?: number
  pageSize?: number
  keyword?: string
  name?: string
  code?: string
  status?: number
}): Promise<
  ApiResponse<{
    list: Dict[]
    total: number
    page: number
    pageSize: number
  }>
> => {
  return service({
    url: '/dict',
    method: 'get',
    params,
  })
}

/**
 * 创建字典
 * 使用场景：字典管理页面新增字典
 */
export const createDict = (data: {
  dictName: string
  dictCode: string
  description: string
  status: number
}): Promise<ApiResponse<Dict>> => {
  return service({
    url: '/dict',
    method: 'post',
    data,
  })
}

/**
 * 更新字典
 * 使用场景：字典管理页面编辑字典、切换字典状态
 */
export const updateDict = (
  id: number,
  data: {
    dictName?: string
    dictCode?: string
    description?: string
    status?: number
  },
): Promise<ApiResponse<Dict>> => {
  return service({
    url: `/dict/${id}`,
    method: 'put',
    data,
  })
}

/**
 * 删除字典
 * 使用场景：字典管理页面删除字典（同时删除关联字典项）
 */
export const deleteDict = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/dict/${id}`,
    method: 'delete',
  })
}

/**
 * 创建字典项
 * 使用场景：字典管理页面添加明细项
 */
export const createDictItem = (
  dictId: number,
  data: {
    itemKey: string
    itemValue: string
    sort: number
    description: string
    status: number
  },
): Promise<ApiResponse<DictItem>> => {
  return service({
    url: '/dict/item',
    method: 'post',
    data: { dictId, ...data },
  })
}

/**
 * 更新字典项
 * 使用场景：字典管理页面编辑明细项、切换明细状态
 */
export const updateDictItem = (
  id: number,
  data: {
    dictId?: number
    itemKey?: string
    itemValue?: string
    sort?: number
    description?: string
    status?: number
  },
): Promise<ApiResponse<DictItem>> => {
  return service({
    url: '/dict/item',
    method: 'put',
    data: { id, ...data },
  })
}

/**
 * 删除字典项
 * 使用场景：字典管理页面删除明细项
 */
export const deleteDictItem = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/dict/item/${id}`,
    method: 'delete',
  })
}

/**
 * 根据字典 ID 获取字典项列表
 * 使用场景：字典管理页面打开明细抽屉时加载明细列表
 */
export const getDictItemsByDictId = (
  dictId: number,
): Promise<ApiResponse<{ list: DictItem[]; total: number }>> => {
  return service({
    url: '/dict/items',
    method: 'get',
    params: { dict_id: dictId },
  })
}

/**
 * 根据字典编码获取启用的字典项
 * 使用场景：全局下拉框、表单选项等需要字典选项的场景
 */
export const getDictItemsByCode = (code: string): Promise<ApiResponse<DictItem[]>> => {
  return service({
    url: `/dict/code/${code}/active-items`,
    method: 'get',
  })
}
