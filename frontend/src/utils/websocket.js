/**
 * WebSocket工具类
 */
export class WebSocketManager {
  constructor(url, options = {}) {
    this.url = url
    this.ws = null
    this.reconnectTimer = null
    this.heartbeatTimer = null
    this.heartbeatTimeoutTimer = null
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = options.maxReconnectAttempts || 10
    this.reconnectInterval = options.reconnectInterval || 5000
    this.heartbeatInterval = options.heartbeatInterval || 60000
    this.heartbeatTimeout = options.heartbeatTimeout || 10000 // 心跳超时时间（10秒）
    this.onMessage = options.onMessage || null
    this.onOpen = options.onOpen || null
    this.onClose = options.onClose || null
    this.onError = options.onError || null
    this.connected = false
    this.lastHeartbeatAck = null // 最后一次收到心跳确认的时间
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
        this.lastHeartbeatAck = Date.now() // 初始化，避免首次心跳误判
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
          if (message.type === 'heartbeat_ack') {
            // 收到心跳确认，更新最后确认时间
            this.lastHeartbeatAck = Date.now()
            this.clearHeartbeatTimeout()
            console.log('收到心跳确认:', message.data?.message || '心跳正常')
            return
          }
          
          // 兼容旧版本的心跳响应
          if (message.type === 'heartbeat' || message.type === 'pong') {
            this.lastHeartbeatAck = Date.now()
            this.clearHeartbeatTimeout()
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
        // 发送心跳
        this.send({ type: 'heartbeat', timestamp: Date.now() })
        
        // 设置心跳超时检测
        this.setHeartbeatTimeout()
      }
    }, this.heartbeatInterval)
  }
  
  /**
   * 设置心跳超时检测
   */
  setHeartbeatTimeout() {
    this.clearHeartbeatTimeout()
    
    const heartbeatSendTime = Date.now()
    
    this.heartbeatTimeoutTimer = setTimeout(() => {
      // 检查是否在超时时间内收到了心跳确认
      if (this.lastHeartbeatAck && this.lastHeartbeatAck >= heartbeatSendTime) {
        // 收到了确认，正常
        return
      }
      
      // 未收到确认，认为连接异常
      console.warn('心跳超时，未收到服务器确认，连接可能异常，尝试重连')
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.close()
      }
    }, this.heartbeatTimeout)
  }
  
  /**
   * 清除心跳超时检测
   */
  clearHeartbeatTimeout() {
    if (this.heartbeatTimeoutTimer) {
      clearTimeout(this.heartbeatTimeoutTimer)
      this.heartbeatTimeoutTimer = null
    }
  }

  /**
   * 停止心跳
   */
  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
    this.clearHeartbeatTimeout()
    this.lastHeartbeatAck = null
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

