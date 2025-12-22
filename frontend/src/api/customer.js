import request from '@/utils/request'

/**
 * 获取客户列表
 * @param {object} params - 查询参数
 */
export const getCustomers = (params) => {
  return request.get('/customers', { params })
}

/**
 * 获取客户详情
 * @param {number} id - 客户ID
 */
export const getCustomer = (id) => {
  return request.get(`/customers/${id}`)
}

/**
 * 更新客户信息
 * @param {number} id - 客户ID
 * @param {object} data - 客户数据
 */
export const updateCustomer = (id, data) => {
  return request.put(`/customers/${id}`, data)
}

/**
 * 删除客户
 * @param {number} id - 客户ID
 */
export const deleteCustomer = (id) => {
  return request.delete(`/customers/${id}`)
}

