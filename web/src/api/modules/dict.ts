import service from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface Dict {
  id: number
  name: string
  code: string
  type: string
  status: number
  createdAt: string
  updatedAt: string
  items?: DictItem[]
}

export interface DictItem {
  id: number
  dictId: number
  label: string
  value: string
  sort: number
  status: number
  createdAt: string
  updatedAt: string
}

// 获取字典列表
export const getDictList = (params?: {
  page?: number
  pageSize?: number
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

// 创建字典
export const createDict = (data: {
  name: string
  code: string
  type: string
  status: number
}): Promise<ApiResponse<Dict>> => {
  return service({
    url: '/dict',
    method: 'post',
    data,
  })
}

// 更新字典
export const updateDict = (
  id: number,
  data: {
    name?: string
    code?: string
    type?: string
    status?: number
  },
): Promise<ApiResponse<Dict>> => {
  return service({
    url: `/dict/${id}`,
    method: 'put',
    data,
  })
}

// 删除字典
export const deleteDict = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/dict/${id}`,
    method: 'delete',
  })
}

// 获取字典项列表
export const getDictItemList = (dictCode: string): Promise<ApiResponse<DictItem[]>> => {
  return service({
    url: `/dict/${dictCode}/items`,
    method: 'get',
  })
}

// 创建字典项
export const createDictItem = (
  dictId: number,
  data: {
    label: string
    value: string
    sort: number
    status: number
  },
): Promise<ApiResponse<DictItem>> => {
  return service({
    url: `/dict/${dictId}/items`,
    method: 'post',
    data,
  })
}

// 更新字典项
export const updateDictItem = (
  id: number,
  data: {
    label?: string
    value?: string
    sort?: number
    status?: number
  },
): Promise<ApiResponse<DictItem>> => {
  return service({
    url: '/dict/item',
    method: 'put',
    data: { id, ...data },
  })
}

// 删除字典项
export const deleteDictItem = (id: number): Promise<ApiResponse> => {
  return service({
    url: `/dict/item/${id}`,
    method: 'delete',
  })
}

// 根据 ID 获取字典
export const getDictById = (id: number): Promise<ApiResponse<Dict>> => {
  return service({
    url: `/dict/${id}`,
    method: 'get',
  })
}

// 根据编码获取字典
export const getDictByCode = (code: string): Promise<ApiResponse<Dict>> => {
  return service({
    url: '/dict/code',
    method: 'get',
    params: { code },
  })
}

// 根据 ID 获取字典项
export const getDictItemById = (id: number): Promise<ApiResponse<DictItem>> => {
  return service({
    url: `/dict/item/${id}`,
    method: 'get',
  })
}

// 根据字典 ID 获取字典项列表
export const getDictItemsByDictId = (dictId: number): Promise<ApiResponse<DictItem[]>> => {
  return service({
    url: '/dict/items',
    method: 'get',
    params: { dictId },
  })
}

// 更新字典项
export const updateDictItemNew = (
  id: number,
  data: {
    label?: string
    value?: string
    sort?: number
    status?: number
  },
): Promise<ApiResponse<DictItem>> => {
  return service({
    url: '/dict/item',
    method: 'put',
    data: { id, ...data },
  })
}
