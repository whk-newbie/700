import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createWebSocket } from '@/utils/websocket'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'

export const useWebSocketStore = defineStore('websocket', () => {
  const wsManager = ref(null)
  const connected = ref(false)
  const reconnectAttempts = ref(0)
  const messageHandlers = ref(new Map()) // 存储消息处理器

  /**
   * 连接WebSocket
   */
  const connect = () => {
    const authStore = useAuthStore()
    const token = authStore.token
    
    if (!token) {
      console.warn('未登录，无法建立WebSocket连接')
      return
    }

    // 如果已连接，先断开
    if (wsManager.value && connected.value) {
      disconnect()
    }

    // 构建WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/api/ws/dashboard?token=${token}`

    wsManager.value = createWebSocket(wsUrl, {
      onMessage: (message) => {
        // 调用所有注册的消息处理器
        messageHandlers.value.forEach((handler) => {
          try {
            handler(message)
          } catch (error) {
            console.error('消息处理器执行失败:', error)
          }
        })
      },
      onOpen: () => {
        connected.value = true
        reconnectAttempts.value = 0
        console.log('WebSocket连接成功')
        ElMessage.success('实时连接已建立')
      },
      onClose: () => {
        connected.value = false
        console.log('WebSocket连接关闭')
        if (reconnectAttempts.value > 0) {
          ElMessage.warning('连接已断开，正在重连...')
        }
      },
      onError: (error) => {
        console.error('WebSocket错误:', error)
        ElMessage.error('连接错误，请检查网络')
      }
    })

    wsManager.value.connect()
  }

  /**
   * 断开WebSocket连接
   */
  const disconnect = () => {
    if (wsManager.value) {
      wsManager.value.disconnect()
      wsManager.value = null
    }
    connected.value = false
    messageHandlers.value.clear()
  }

  /**
   * 注册消息处理器
   * @param {string} id - 处理器ID
   * @param {function} handler - 处理函数
   */
  const registerMessageHandler = (id, handler) => {
    messageHandlers.value.set(id, handler)
  }

  /**
   * 取消注册消息处理器
   * @param {string} id - 处理器ID
   */
  const unregisterMessageHandler = (id) => {
    messageHandlers.value.delete(id)
  }

  /**
   * 发送消息
   * @param {object} message - 消息对象
   */
  const send = (message) => {
    if (wsManager.value && connected.value) {
      wsManager.value.send(message)
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  return {
    connected,
    reconnectAttempts,
    connect,
    disconnect,
    registerMessageHandler,
    unregisterMessageHandler,
    send
  }
})

