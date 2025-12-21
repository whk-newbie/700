import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getCurrentUser } from '@/api/auth'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const loading = ref(false)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)
  const userRole = computed(() => user.value?.role || '')
  const isAdmin = computed(() => userRole.value === 'admin')
  const isUser = computed(() => userRole.value === 'user')
  const isSubAccount = computed(() => userRole.value === 'subaccount')

  // 登录
  const login = async (credentials) => {
    loading.value = true
    try {
      const response = await apiLogin(credentials)
      const { data } = response

      // 保存token
      token.value = data.token
      localStorage.setItem('token', data.token)

      // 保存用户信息
      user.value = data.user
      localStorage.setItem('user', JSON.stringify(data.user))

      // 跳转到首页
      router.push('/dashboard')

      return response
    } catch (error) {
      throw error
    } finally {
      loading.value = false
    }
  }

  // 子账号登录
  const subAccountLogin = async (credentials) => {
    loading.value = true
    try {
      const response = await apiLogin(credentials)
      const { data } = response

      // 保存token
      token.value = data.token
      localStorage.setItem('token', data.token)

      // 保存用户信息（子账号信息）
      user.value = {
        ...data.user,
        role: 'subaccount',
        groupId: data.group_id,
        activationCode: credentials.activation_code
      }
      localStorage.setItem('user', JSON.stringify(user.value))

      // 跳转到首页
      router.push('/dashboard')

      return response
    } catch (error) {
      throw error
    } finally {
      loading.value = false
    }
  }

  // 获取当前用户信息
  const fetchUser = async () => {
    if (!token.value) return

    try {
      const response = await getCurrentUser()
      user.value = response.data
      localStorage.setItem('user', JSON.stringify(response.data))
    } catch (error) {
      // Token可能已过期
      logout()
      throw error
    }
  }

  // 检查认证状态
  const checkAuth = async () => {
    if (token.value && !user.value) {
      await fetchUser()
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    router.push('/login')
  }

  // 更新用户信息
  const updateUser = (userData) => {
    user.value = { ...user.value, ...userData }
    localStorage.setItem('user', JSON.stringify(user.value))
  }

  // 刷新token
  const refreshToken = async () => {
    try {
      // 这里应该调用刷新token的API
      // const response = await refreshTokenAPI()
      // token.value = response.data.token
      // localStorage.setItem('token', response.data.token)
    } catch (error) {
      logout()
      throw error
    }
  }

  return {
    // 状态
    token,
    user,
    loading,

    // 计算属性
    isAuthenticated,
    userRole,
    isAdmin,
    isUser,
    isSubAccount,

    // 方法
    login,
    subAccountLogin,
    fetchUser,
    checkAuth,
    logout,
    updateUser,
    refreshToken
  }
})
