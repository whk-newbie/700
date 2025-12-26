import JSEncrypt from 'jsencrypt'

/**
 * 使用RSA公钥加密数据
 * @param {string} plaintext - 要加密的明文
 * @param {string} publicKeyPEM - PEM格式的RSA公钥
 * @returns {Promise<string>} Base64编码的密文
 */
export function encryptWithRSA(plaintext, publicKeyPEM) {
  return new Promise((resolve, reject) => {
    try {
      const encrypt = new JSEncrypt({ default_key_size: 2048 })
      encrypt.setPublicKey(publicKeyPEM)

      // 使用PKCS1填充（与后端匹配）
      const encrypted = encrypt.encrypt(plaintext)
      if (!encrypted) {
        reject(new Error('RSA加密失败'))
        return
      }

      resolve(encrypted)
    } catch (error) {
      reject(error)
    }
  })
}

/**
 * 验证RSA公钥格式
 * @param {string} publicKeyPEM - PEM格式的RSA公钥
 * @returns {boolean}
 */
export function validatePublicKey(publicKeyPEM) {
  if (!publicKeyPEM || typeof publicKeyPEM !== 'string') {
    return false
  }
  
  // 检查是否包含PUBLIC KEY标记
  return publicKeyPEM.includes('-----BEGIN PUBLIC KEY-----') && 
         publicKeyPEM.includes('-----END PUBLIC KEY-----')
}

