import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/utils/request'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isSubAccount = computed(() => user.value?.role === 'subaccount')

  // 登录
  const login = async (username, password) => {
    try {
      const res = await request.post('/auth/login', {
        username,
        password
      })
      
      if (res.code === 200 && res.data) {
        token.value = res.data.token
        user.value = res.data.user || res.data // 兼容不同的响应格式
        
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('user', JSON.stringify(user.value))
        
        return { success: true }
      } else {
        return { success: false, message: res.message || '登录失败' }
      }
    } catch (error) {
      return { success: false, message: error.message || '登录失败' }
    }
  }

  // 子账号登录
  const loginSubAccount = async (activationCode, password) => {
    try {
      const res = await request.post('/auth/login-subaccount', {
        activation_code: activationCode,
        password
      })
      
      if (res.code === 200 && res.data) {
        token.value = res.data.token
        // 子账号登录返回的是group信息，需要转换为user格式
        if (res.data.group) {
          user.value = {
            id: res.data.group.id,
            activation_code: res.data.group.activation_code,
            category: res.data.group.category,
            role: 'subaccount'
          }
        } else {
          user.value = res.data
        }
        
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('user', JSON.stringify(user.value))
        
        return { success: true }
      } else {
        return { success: false, message: res.message || '登录失败' }
      }
    } catch (error) {
      return { success: false, message: error.message || '登录失败' }
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // 获取当前用户信息
  const fetchUserInfo = async () => {
    try {
      const res = await request.get('/auth/me')
      if (res.code === 200) {
        user.value = res.data
        localStorage.setItem('user', JSON.stringify(res.data))
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }

  // 检查认证状态
  const checkAuth = () => {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    
    if (storedToken) {
      token.value = storedToken
      if (storedUser) {
        try {
          user.value = JSON.parse(storedUser)
        } catch (e) {
          console.error('解析用户信息失败:', e)
          user.value = null
        }
      }
    } else {
      token.value = ''
      user.value = null
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    isSubAccount,
    login,
    loginSubAccount,
    logout,
    fetchUserInfo,
    checkAuth
  }
})

