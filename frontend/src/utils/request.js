import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import router from '@/router'

// 创建axios实例
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const authStore = useAuthStore()

    // 添加认证token
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }

    // 如果是FormData，不设置Content-Type，让浏览器自动设置
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type']
    }

    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const { data } = response

    // 请求成功
    if (data.code === 1000) {
      return data
    }

    // 处理业务错误
    handleBusinessError(data)
    return Promise.reject(new Error(data.message || '请求失败'))
  },
  error => {
    const { response } = error

    if (response) {
      // 服务器返回了错误状态码
      const { status, data } = response

      switch (status) {
        case 401:
          handleUnauthorized()
          break
        case 403:
          ElMessage.error('权限不足')
          break
        case 404:
          ElMessage.error('请求地址不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          if (data && data.message) {
            ElMessage.error(data.message)
          } else {
            ElMessage.error(`请求失败 (${status})`)
          }
      }
    } else if (error.code === 'ECONNABORTED') {
      // 请求超时
      ElMessage.error('请求超时，请检查网络连接')
    } else {
      // 网络错误
      ElMessage.error('网络错误，请检查网络连接')
    }

    return Promise.reject(error)
  }
)

// 处理业务错误
function handleBusinessError(data) {
  const { code, message } = data

  switch (code) {
    case 2001: // 未登录
    case 2002: // Token无效
    case 2003: // Token过期
      handleUnauthorized()
      break
    case 2007: // 权限不足
      ElMessage.error('权限不足')
      break
    default:
      ElMessage.error(message || '操作失败')
  }
}

// 处理未授权
function handleUnauthorized() {
  const authStore = useAuthStore()

  // 清除认证信息
  authStore.logout()

  // 显示重新登录提示
  ElMessageBox.confirm(
    '登录已过期，请重新登录',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
      showCancelButton: false
    }
  ).then(() => {
    router.push('/login')
  })
}

// 导出请求方法
export default request

// HTTP方法封装
export const http = {
  get: (url, params = {}, config = {}) => request.get(url, { ...config, params }),
  post: (url, data = {}, config = {}) => request.post(url, data, config),
  put: (url, data = {}, config = {}) => request.put(url, data, config),
  patch: (url, data = {}, config = {}) => request.patch(url, data, config),
  delete: (url, config = {}) => request.delete(url, config),
  upload: (url, formData, config = {}) => request.post(url, formData, {
    ...config,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
