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
 */
export const generateQRCode = (id) => {
  return request.post(`/line-accounts/${id}/generate-qr`)
}

