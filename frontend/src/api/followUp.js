import request from '@/utils/request'

/**
 * 获取跟进记录列表
 * @param {object} params - 查询参数
 */
export const getFollowUps = (params) => {
  return request.get('/follow-ups', { params })
}

/**
 * 获取跟进记录详情
 * @param {number} id - 记录ID
 */
export const getFollowUp = (id) => {
  return request.get(`/follow-ups/${id}`)
}

/**
 * 创建跟进记录
 * @param {object} data - 记录数据
 */
export const createFollowUp = (data) => {
  return request.post('/follow-ups', data)
}

/**
 * 更新跟进记录
 * @param {number} id - 记录ID
 * @param {object} data - 记录数据
 */
export const updateFollowUp = (id, data) => {
  return request.put(`/follow-ups/${id}`, data)
}

/**
 * 删除跟进记录
 * @param {number} id - 记录ID
 */
export const deleteFollowUp = (id) => {
  return request.delete(`/follow-ups/${id}`)
}

/**
 * 批量创建跟进记录
 * @param {array} data - 记录数据数组
 */
export const batchCreateFollowUp = (data) => {
  return request.post('/follow-ups/batch', data)
}

