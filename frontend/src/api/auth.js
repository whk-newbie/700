import request from '@/utils/request'

/**
 * 用户登录
 * @param {string} username - 用户名
 * @param {string} password - 密码
 */
export const login = (username, password) => {
  return request.post('/auth/login', {
    username,
    password
  })
}

/**
 * 子账号登录
 * @param {string} activationCode - 激活码
 * @param {string} password - 密码
 */
export const loginSubAccount = (activationCode, password) => {
  return request.post('/auth/login-subaccount', {
    activation_code: activationCode,
    password
  })
}

/**
 * 登出
 */
export const logout = () => {
  return request.post('/auth/logout')
}

/**
 * 获取当前用户信息
 */
export const getCurrentUser = () => {
  return request.get('/auth/me')
}

/**
 * 刷新Token
 */
export const refreshToken = () => {
  return request.post('/auth/refresh')
}

