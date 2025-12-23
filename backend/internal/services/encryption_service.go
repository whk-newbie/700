package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"line-management/pkg/logger"
)

// EncryptionService 加密服务
type EncryptionService struct {
	key []byte
}

var encryptionService *EncryptionService

// GetEncryptionService 获取加密服务实例（单例）
func GetEncryptionService() *EncryptionService {
	if encryptionService == nil {
		key := getEncryptionKey()
		encryptionService = &EncryptionService{
			key: key,
		}
	}
	return encryptionService
}

// getEncryptionKey 获取加密密钥（从环境变量或使用默认值）
func getEncryptionKey() []byte {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		// 默认密钥（生产环境必须设置环境变量）
		keyStr = "default-encryption-key-32-bytes-long!!"
		logger.Warnf("使用默认加密密钥，生产环境请设置ENCRYPTION_KEY环境变量")
	}

	// 确保密钥长度为32字节（AES-256）
	key := []byte(keyStr)
	if len(key) < 32 {
		// 如果密钥太短，填充到32字节
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		// 如果密钥太长，截取前32字节
		key = key[:32]
	}

	return key
}

// Encrypt 加密字符串
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Base64解码
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < nonceSize {
		return "", errors.New("密文太短")
	}

	// 提取nonce和密文
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

