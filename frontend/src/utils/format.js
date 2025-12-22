import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

dayjs.locale('zh-cn')

/**
 * 格式化日期时间
 * @param {string|Date} date - 日期
 * @param {string} format - 格式，默认 'YYYY-MM-DD HH:mm:ss'
 */
export const formatDateTime = (date, format = 'YYYY-MM-DD HH:mm:ss') => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化日期
 * @param {string|Date} date - 日期
 * @param {string} format - 格式，默认 'YYYY-MM-DD'
 */
export const formatDate = (date, format = 'YYYY-MM-DD') => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化时间
 * @param {string|Date} date - 日期
 * @param {string} format - 格式，默认 'HH:mm:ss'
 */
export const formatTime = (date, format = 'HH:mm:ss') => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 相对时间（如：3分钟前）
 * @param {string|Date} date - 日期
 */
export const formatRelativeTime = (date) => {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 */
export const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

/**
 * 格式化数字（千分位）
 * @param {number} num - 数字
 */
export const formatNumber = (num) => {
  if (num === null || num === undefined) return '-'
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 格式化百分比
 * @param {number} num - 数字
 * @param {number} decimals - 小数位数，默认2
 */
export const formatPercent = (num, decimals = 2) => {
  if (num === null || num === undefined) return '-'
  return (num * 100).toFixed(decimals) + '%'
}

/**
 * 格式化手机号（中间4位隐藏）
 * @param {string} phone - 手机号
 */
export const formatPhone = (phone) => {
  if (!phone) return '-'
  if (phone.length !== 11) return phone
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

/**
 * 格式化状态文本
 * @param {string|number} status - 状态值
 * @param {object} statusMap - 状态映射对象
 */
export const formatStatus = (status, statusMap) => {
  return statusMap[status] || status
}

/**
 * 截断文本
 * @param {string} text - 文本
 * @param {number} length - 最大长度
 * @param {string} suffix - 后缀，默认 '...'
 */
export const truncateText = (text, length = 50, suffix = '...') => {
  if (!text) return '-'
  if (text.length <= length) return text
  return text.substring(0, length) + suffix
}

