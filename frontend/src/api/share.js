import request from '@/utils/request'

/**
 * 创建分组分享
 * @param {number} groupId 分组ID
 * @returns {Promise}
 */
export function createGroupShare(groupId) {
  return request({
    url: `/groups/${groupId}/share`,
    method: 'post'
  })
}

/**
 * 获取分组的分享信息
 * @param {number} groupId 分组ID
 * @returns {Promise}
 */
export function getGroupShare(groupId) {
  return request({
    url: `/groups/${groupId}/share`,
    method: 'get'
  })
}

/**
 * 删除分组分享
 * @param {number} groupId 分组ID
 * @returns {Promise}
 */
export function deleteGroupShare(groupId) {
  return request({
    url: `/groups/${groupId}/share`,
    method: 'delete'
  })
}

/**
 * 获取分享信息（公开接口，不需要认证）
 * @param {string} code 分享码
 * @returns {Promise}
 */
export function getShareInfo(code) {
  return request({
    url: '/share/info',
    method: 'get',
    params: { code }
  })
}

/**
 * 验证分享密码（公开接口，不需要认证）
 * @param {string} code 分享码
 * @param {string} password 密码
 * @returns {Promise}
 */
export function verifySharePassword(code, password) {
  return request({
    url: '/share/verify',
    method: 'post',
    data: { code, password }
  })
}

