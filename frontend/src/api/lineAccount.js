import request from '@/utils/request'

/**
 * 获取Line账号列表
 * @param {object} params - 查询参数
 */
export const getLineAccounts = (params) => {
  return request.get('/line-accounts', { params })
}

/**
 * 获取Line账号详情
 * @param {number} id - 账号ID
 */
export const getLineAccount = (id) => {
  return request.get(`/line-accounts/${id}`)
}

/**
 * 创建Line账号
 * @param {object} data - 账号数据
 */
export const createLineAccount = (data) => {
  return request.post('/line-accounts', data)
}

/**
 * 更新Line账号
 * @param {number} id - 账号ID
 * @param {object} data - 账号数据
 */
export const updateLineAccount = (id, data) => {
  return request.put(`/line-accounts/${id}`, data)
}

/**
 * 删除Line账号
 * @param {number} id - 账号ID
 */
export const deleteLineAccount = (id) => {
  return request.delete(`/line-accounts/${id}`)
}

/**
 * 生成二维码
 * @param {number} id - 账号ID
 * @param {string} content - 二维码内容（可选）
 */
export const generateQRCode = (id, content) => {
  const params = content ? { content } : {}
  return request.post(`/line-accounts/${id}/generate-qr`, null, { params })
}

/**
 * 批量删除Line账号
 * @param {number[]} ids - 账号ID数组
 */
export const batchDeleteLineAccounts = (ids) => {
  return request.post('/line-accounts/batch/delete', { ids })
}

/**
 * 批量更新Line账号状态（强制下线）
 * @param {object} data - 批量更新数据 { ids: [], online_status: string }
 */
export const batchUpdateLineAccounts = (data) => {
  return request.post('/line-accounts/batch/update', data)
}

