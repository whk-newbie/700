package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"line-management/pkg/logger"
)

// RSAService RSA加密服务
type RSAService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mu         sync.RWMutex
}

var rsaService *RSAService
var rsaOnce sync.Once

// GetRSAService 获取RSA服务实例（单例）
func GetRSAService() *RSAService {
	rsaOnce.Do(func() {
		rsaService = &RSAService{}
		rsaService.loadOrGenerateKeys()
	})
	return rsaService
}

// loadOrGenerateKeys 加载或生成RSA密钥对
func (s *RSAService) loadOrGenerateKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 尝试从文件加载密钥
	// 优先使用环境变量指定的路径，否则使用默认路径
	keysDir := os.Getenv("RSA_KEYS_DIR")
	if keysDir == "" {
		keysDir = "keys"
	}
	
	privateKeyPath := filepath.Join(keysDir, "rsa_private.pem")
	publicKeyPath := filepath.Join(keysDir, "rsa_public.pem")

	// 创建keys目录
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		logger.Errorf("创建keys目录失败: %v", err)
	}

	// 尝试加载私钥
	if privateKey, err := s.loadPrivateKey(privateKeyPath); err == nil {
		s.privateKey = privateKey
		s.publicKey = &privateKey.PublicKey
		logger.Infof("成功加载RSA密钥对")
		return
	}

	// 如果加载失败，生成新的密钥对
	logger.Infof("未找到RSA密钥对，正在生成新的密钥对...")
	if err := s.generateKeys(privateKeyPath, publicKeyPath); err != nil {
		logger.Errorf("生成RSA密钥对失败: %v", err)
		// 如果生成失败，使用临时密钥（不推荐，但保证服务可用）
		// 注意：临时密钥在容器重启后会变化，建议检查keys目录权限
		s.generateTemporaryKeys()
		logger.Warnf("使用临时RSA密钥对，请检查keys目录权限: %s", keysDir)
	}
}

// loadPrivateKey 从文件加载私钥
func (s *RSAService) loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("无法解析PEM块")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("不是RSA私钥")
		}
		return rsaKey, nil
	}

	return privateKey, nil
}

// generateKeys 生成RSA密钥对并保存到文件
func (s *RSAService) generateKeys(privateKeyPath, publicKeyPath string) error {
	// 生成2048位RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 保存私钥
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return err
	}

	// 保存公钥
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return err
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	logger.Infof("RSA密钥对已生成并保存到 %s 和 %s", privateKeyPath, publicKeyPath)
	return nil
}

// generateTemporaryKeys 生成临时密钥对（不保存）
func (s *RSAService) generateTemporaryKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logger.Errorf("生成临时RSA密钥对失败: %v", err)
		return
	}
	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey
	logger.Warnf("使用临时RSA密钥对，重启后密钥会变化")
}

// GetPublicKeyPEM 获取公钥的PEM格式字符串
func (s *RSAService) GetPublicKeyPEM() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.publicKey == nil {
		return "", errors.New("公钥未初始化")
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// Decrypt 使用私钥解密数据（Base64编码的密文）
func (s *RSAService) Decrypt(encryptedData string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.privateKey == nil {
		return "", errors.New("私钥未初始化")
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	// RSA解密（使用OAEP填充）
	plaintext, err := rsa.DecryptOAEP(
		rand.Reader,
		s.privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

