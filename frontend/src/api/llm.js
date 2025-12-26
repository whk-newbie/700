import request from '@/utils/request'
import { encryptWithRSA, validatePublicKey } from '@/utils/rsa'

/**
 * 获取RSA公钥（用于加密API Key）
 */
export const getRSAPublicKey = () => {
  return request.get('/admin/llm/rsa-public-key')
}

/**
 * 获取OpenAI API Key配置（管理员）
 */
export const getOpenAIAPIKey = () => {
  return request.get('/admin/llm/openai-key')
}

/**
 * 更新OpenAI API Key（管理员）
 * API Key会在前端使用RSA公钥加密后传输
 * @param {string} apiKey - 明文的API Key
 * @returns {Promise} 返回更新结果
 */
export const updateOpenAIAPIKey = async (apiKey) => {
  try {
    // 1. 获取RSA公钥
    const publicKeyResponse = await getRSAPublicKey()
    // 响应格式可能是 { data: { data: { public_key: ... } } } 或 { data: { public_key: ... } }
    const publicKeyPEM = publicKeyResponse.data?.data?.public_key || publicKeyResponse.data?.public_key
    
    if (!publicKeyPEM || !validatePublicKey(publicKeyPEM)) {
      throw new Error('获取的RSA公钥格式无效')
    }
    
    // 2. 使用RSA公钥加密API Key
    const encryptedAPIKey = await encryptWithRSA(apiKey, publicKeyPEM)
    
    // 3. 发送加密后的API Key
    return request.put('/admin/llm/openai-key', {
      encrypted_api_key: encryptedAPIKey
    })
  } catch (error) {
    console.error('更新OpenAI API Key失败:', error)
    throw error
  }
}

/**
 * OpenAI API转发接口
 * @param {object} data - OpenAI API请求数据（格式与OpenAI文档一致）
 */
export const proxyOpenAIAPI = (data) => {
  return request.post('/llm/proxy/openai', data)
}

/**
 * 获取LLM调用日志列表
 * @param {object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {number} params.config_id - 配置ID（可选）
 * @param {number} params.template_id - 模板ID（可选）
 * @param {number} params.group_id - 分组ID（可选）
 * @param {string} params.activation_code - 激活码（可选）
 * @param {string} params.status - 状态：success/error（可选）
 * @param {string} params.start_time - 开始时间（可选）
 * @param {string} params.end_time - 结束时间（可选）
 */
export const getLLMCallLogs = (params) => {
  return request.get('/admin/llm/call-logs', { params })
}

