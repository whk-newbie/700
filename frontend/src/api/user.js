import request from '@/utils/request'

/**
 * 获取用户列表（管理员）
 * @param {object} params - 查询参数
 */
export const getUsers = (params) => {
  return request.get('/admin/users', { params })
}

/**
 * 创建用户（管理员）
 * @param {object} data - 用户数据
 */
export const createUser = (data) => {
  return request.post('/admin/users', data)
}

/**
 * 更新用户（管理员）
 * @param {number} id - 用户ID
 * @param {object} data - 用户数据
 */
export const updateUser = (id, data) => {
  return request.put(`/admin/users/${id}`, data)
}

/**
 * 删除用户（管理员）
 * @param {number} id - 用户ID
 */
export const deleteUser = (id) => {
  return request.delete(`/admin/users/${id}`)
}

