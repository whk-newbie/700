import request from '@/utils/request'

/**
 * 获取分组列表
 * @param {object} params - 查询参数
 */
export const getGroups = (params) => {
  return request.get('/groups', { params })
}

/**
 * 获取分组详情
 * @param {number} id - 分组ID
 */
export const getGroup = (id) => {
  return request.get(`/groups/${id}`)
}

/**
 * 创建分组
 * @param {object} data - 分组数据
 */
export const createGroup = (data) => {
  return request.post('/groups', data)
}

/**
 * 更新分组
 * @param {number} id - 分组ID
 * @param {object} data - 分组数据
 */
export const updateGroup = (id, data) => {
  return request.put(`/groups/${id}`, data)
}

/**
 * 删除分组
 * @param {number} id - 分组ID
 */
export const deleteGroup = (id) => {
  return request.delete(`/groups/${id}`)
}

/**
 * 重新生成激活码
 * @param {number} id - 分组ID
 */
export const regenerateCode = (id) => {
  return request.post(`/groups/${id}/regenerate-code`)
}

/**
 * 获取分组分类列表
 */
export const getGroupCategories = () => {
  return request.get('/groups/categories')
}

/**
 * 批量删除分组
 * @param {number[]} ids - 分组ID数组
 */
export const batchDeleteGroups = (ids) => {
  return request.post('/groups/batch/delete', { ids })
}

/**
 * 批量更新分组
 * @param {object} data - 批量更新数据 { ids: [], is_active?: boolean, category?: string, dedup_scope?: string }
 */
export const batchUpdateGroups = (data) => {
  return request.post('/groups/batch/update', data)
}

