/**
 * WebSocket工具类
 */
export class WebSocketManager {
  constructor(url, options = {}) {
    this.url = url
    this.ws = null
    this.reconnectTimer = null
    this.heartbeatTimer = null
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = options.maxReconnectAttempts || 10
    this.reconnectInterval = options.reconnectInterval || 5000
    this.heartbeatInterval = options.heartbeatInterval || 60000
    this.onMessage = options.onMessage || null
    this.onOpen = options.onOpen || null
    this.onClose = options.onClose || null
    this.onError = options.onError || null
    this.connected = false
  }

  /**
   * 连接WebSocket
   */
  connect() {
    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.connected = true
        this.reconnectAttempts = 0
        console.log('WebSocket连接成功')
        
        // 启动心跳
        this.startHeartbeat()
        
        if (this.onOpen) {
          this.onOpen()
        }
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          
          // 处理心跳响应
          if (message.type === 'heartbeat' || message.type === 'pong') {
            return
          }
          
          if (this.onMessage) {
            this.onMessage(message)
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        if (this.onError) {
          this.onError(error)
        }
      }

      this.ws.onclose = () => {
        this.connected = false
        this.stopHeartbeat()
        console.log('WebSocket连接关闭')
        
        if (this.onClose) {
          this.onClose()
        }
        
        // 自动重连
        this.reconnect()
      }
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      this.reconnect()
    }
  }

  /**
   * 发送消息
   * @param {object} message - 消息对象
   */
  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  /**
   * 启动心跳
   */
  startHeartbeat() {
    this.stopHeartbeat()
    
    this.heartbeatTimer = setInterval(() => {
      if (this.connected) {
        this.send({ type: 'heartbeat', timestamp: Date.now() })
      }
    }, this.heartbeatInterval)
  }

  /**
   * 停止心跳
   */
  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  /**
   * 重连
   */
  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('WebSocket重连次数已达上限')
      return
    }

    if (this.reconnectTimer) {
      return
    }

    this.reconnectAttempts++
    console.log(`WebSocket将在${this.reconnectInterval / 1000}秒后重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, this.reconnectInterval)
  }

  /**
   * 断开连接
   */
  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    this.stopHeartbeat()
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    
    this.connected = false
  }

  /**
   * 获取连接状态
   */
  isConnected() {
    return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

/**
 * 创建WebSocket连接
 * @param {string} url - WebSocket地址
 * @param {object} options - 配置选项
 */
export const createWebSocket = (url, options = {}) => {
  return new WebSocketManager(url, options)
}

