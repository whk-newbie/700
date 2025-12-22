import request from '@/utils/request'

/**
 * 获取分组统计
 * @param {number} id - 分组ID
 */
export const getGroupStats = (id) => {
  return request.get(`/stats/group/${id}`)
}

/**
 * 获取账号统计
 * @param {number} id - 账号ID
 */
export const getAccountStats = (id) => {
  return request.get(`/stats/account/${id}`)
}

/**
 * 获取总览统计
 */
export const getOverviewStats = () => {
  return request.get('/stats/overview')
}

