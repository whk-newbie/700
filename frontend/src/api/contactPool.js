import request from '@/utils/request'

/**
 * 获取底库统计汇总
 */
export const getContactPoolSummary = () => {
  return request.get('/contact-pool/summary')
}

/**
 * 获取底库列表（按激活码+平台）
 * @param {object} params - 查询参数
 */
export const getContactPoolList = (params) => {
  return request.get('/contact-pool/list', { params })
}

/**
 * 获取底库详细列表
 * @param {object} params - 查询参数
 */
export const getContactPoolDetail = (params) => {
  return request.get('/contact-pool/detail', { params })
}

/**
 * 导入原始联系人
 * @param {FormData} formData - 文件表单数据
 */
export const importContacts = (formData) => {
  return request.post('/contact-pool/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 获取导入批次列表
 * @param {object} params - 查询参数
 */
export const getImportBatches = (params) => {
  return request.get('/contact-pool/import-batches', { params })
}

