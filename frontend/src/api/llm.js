import request from '@/utils/request'

/**
 * 获取大模型配置列表（管理员）
 */
export const getLLMConfigs = () => {
  return request.get('/admin/llm/configs')
}

/**
 * 创建大模型配置（管理员）
 * @param {object} data - 配置数据
 */
export const createLLMConfig = (data) => {
  return request.post('/admin/llm/configs', data)
}

/**
 * 更新大模型配置（管理员）
 * @param {number} id - 配置ID
 * @param {object} data - 配置数据
 */
export const updateLLMConfig = (id, data) => {
  return request.put(`/admin/llm/configs/${id}`, data)
}

/**
 * 删除大模型配置（管理员）
 * @param {number} id - 配置ID
 */
export const deleteLLMConfig = (id) => {
  return request.delete(`/admin/llm/configs/${id}`)
}

/**
 * 测试大模型连接（管理员）
 * @param {number} id - 配置ID
 */
export const testLLMConfig = (id) => {
  return request.post(`/admin/llm/configs/${id}/test`)
}

/**
 * 获取Prompt模板列表（管理员）
 */
export const getLLMTemplates = () => {
  return request.get('/admin/llm/templates')
}

/**
 * 创建Prompt模板（管理员）
 * @param {object} data - 模板数据
 */
export const createLLMTemplate = (data) => {
  return request.post('/admin/llm/templates', data)
}

/**
 * 更新Prompt模板（管理员）
 * @param {number} id - 模板ID
 * @param {object} data - 模板数据
 */
export const updateLLMTemplate = (id, data) => {
  return request.put(`/admin/llm/templates/${id}`, data)
}

/**
 * 删除Prompt模板（管理员）
 * @param {number} id - 模板ID
 */
export const deleteLLMTemplate = (id) => {
  return request.delete(`/admin/llm/templates/${id}`)
}

/**
 * 获取可用配置（Windows客户端）
 */
export const getAvailableConfigs = () => {
  return request.get('/llm/configs')
}

/**
 * 调用大模型
 * @param {object} data - 调用数据
 */
export const callLLM = (data) => {
  return request.post('/llm/call', data)
}

/**
 * 使用模板调用大模型
 * @param {object} data - 调用数据
 */
export const callLLMWithTemplate = (data) => {
  return request.post('/llm/call-template', data)
}

/**
 * 获取模板列表（Windows客户端）
 */
export const getTemplates = () => {
  return request.get('/llm/templates')
}

