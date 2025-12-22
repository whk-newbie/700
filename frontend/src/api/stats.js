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

/**
 * 获取分组进线趋势
 * @param {number} id - 分组ID
 * @param {number} days - 天数（默认7天，最多30天）
 */
export const getGroupIncomingTrend = (id, days = 7) => {
  return request.get(`/stats/group/${id}/trend`, { params: { days } })
}

/**
 * 获取账号进线趋势
 * @param {number} id - 账号ID
 * @param {number} days - 天数（默认7天，最多30天）
 */
export const getAccountIncomingTrend = (id, days = 7) => {
  return request.get(`/stats/account/${id}/trend`, { params: { days } })
}

/**
 * 获取进线日志列表
 * @param {object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {number} params.group_id - 分组ID
 * @param {number} params.line_account_id - 账号ID
 * @param {boolean} params.is_duplicate - 是否重复
 * @param {string} params.start_time - 开始时间（ISO 8601格式）
 * @param {string} params.end_time - 结束时间（ISO 8601格式）
 * @param {string} params.search - 搜索（进线Line ID或显示名称）
 */
export const getIncomingLogs = (params) => {
  return request.get('/stats/incoming-logs', { params })
}

