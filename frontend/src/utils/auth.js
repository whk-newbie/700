/**
 * 认证相关工具函数
 */

/**
 * 获取Token
 */
export const getToken = () => {
  return localStorage.getItem('token') || ''
}

/**
 * 设置Token
 * @param {string} token - Token值
 */
export const setToken = (token) => {
  localStorage.setItem('token', token)
}

/**
 * 移除Token
 */
export const removeToken = () => {
  localStorage.removeItem('token')
}

/**
 * 获取用户信息
 */
export const getUser = () => {
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  
  try {
    return JSON.parse(userStr)
  } catch (e) {
    console.error('解析用户信息失败:', e)
    return null
  }
}

/**
 * 设置用户信息
 * @param {object} user - 用户信息
 */
export const setUser = (user) => {
  localStorage.setItem('user', JSON.stringify(user))
}

/**
 * 移除用户信息
 */
export const removeUser = () => {
  localStorage.removeItem('user')
}

/**
 * 清除所有认证信息
 */
export const clearAuth = () => {
  removeToken()
  removeUser()
}

/**
 * 检查是否已登录
 */
export const isAuthenticated = () => {
  return !!getToken()
}

/**
 * 检查是否为管理员
 */
export const isAdmin = () => {
  const user = getUser()
  return user?.role === 'admin'
}

/**
 * 检查是否为子账号
 */
export const isSubAccount = () => {
  const user = getUser()
  return user?.role === 'subaccount'
}

