import { http } from '@/utils/request'

/**
 * 用户登录
 * @param {Object} data - 登录数据
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码
 */
export function login(data) {
  return http.post('/auth/login', data)
}

/**
 * 子账号登录
 * @param {Object} data - 登录数据
 * @param {string} data.activation_code - 激活码
 * @param {string} data.password - 密码
 */
export function subAccountLogin(data) {
  return http.post('/auth/login-subaccount', data)
}

/**
 * 用户登出
 */
export function logout() {
  return http.post('/auth/logout')
}

/**
 * 获取当前用户信息
 */
export function getCurrentUser() {
  return http.get('/auth/me')
}

/**
 * 刷新Token
 */
export function refreshToken() {
  return http.post('/auth/refresh')
}

/**
 * 修改密码
 * @param {Object} data - 密码数据
 * @param {string} data.old_password - 旧密码
 * @param {string} data.new_password - 新密码
 */
export function changePassword(data) {
  return http.put('/auth/change-password', data)
}

/**
 * 重置子账号密码（管理员）
 * @param {number} userId - 用户ID
 * @param {Object} data - 密码数据
 * @param {string} data.new_password - 新密码
 */
export function resetSubAccountPassword(userId, data) {
  return http.put(`/admin/users/${userId}/reset-password`, data)
}
