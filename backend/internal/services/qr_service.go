package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// QRService 二维码服务
type QRService struct {
	db            *gorm.DB
	staticDir     string
	qrcodeSubDir  string
}

// NewQRService 创建二维码服务实例
func NewQRService() *QRService {
	staticDir := "static"
	qrcodeSubDir := "qrcodes"
	
	// 确保目录存在
	qrcodeDir := filepath.Join(staticDir, qrcodeSubDir)
	if err := os.MkdirAll(qrcodeDir, 0755); err != nil {
		logger.Warnf("创建二维码目录失败: %v", err)
	}

	return &QRService{
		db:           database.GetDB(),
		staticDir:   staticDir,
		qrcodeSubDir: qrcodeSubDir,
	}
}

// GenerateQRCode 为Line账号生成二维码
// content: 二维码内容（如果为空，则自动生成Line添加好友链接）
func (s *QRService) GenerateQRCode(accountID uint, content string) (string, error) {
	// 获取账号信息
	var account models.LineAccount
	if err := s.db.Where("id = ? AND deleted_at IS NULL", accountID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("账号不存在")
		}
		return "", err
	}

	// 如果content为空，自动生成Line添加好友链接
	// Line添加好友链接格式: https://line.me/ti/p/~{line_id}
	if content == "" {
		content = fmt.Sprintf("https://line.me/ti/p/~%s", account.LineID)
	}

	// 生成二维码文件名
	filename := fmt.Sprintf("qr_%d_%s.png", accountID, account.LineID)
	filePath := filepath.Join(s.staticDir, s.qrcodeSubDir, filename)

	// 生成二维码（使用中等错误恢复级别，大小256x256）
	err := qrcode.WriteFile(content, qrcode.Medium, 256, filePath)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %v", err)
	}

	// 更新账号的二维码路径（相对路径，用于前端访问）
	qrCodePath := filepath.Join("/", s.qrcodeSubDir, filename)
	account.QRCodePath = qrCodePath
	if err := s.db.Save(&account).Error; err != nil {
		logger.Warnf("更新账号二维码路径失败: %v", err)
		// 不影响返回结果
	}

	return qrCodePath, nil
}

// GetQRCodePath 获取账号的二维码路径
func (s *QRService) GetQRCodePath(accountID uint) (string, error) {
	var account models.LineAccount
	if err := s.db.Where("id = ? AND deleted_at IS NULL", accountID).
		Select("qr_code_path").First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("账号不存在")
		}
		return "", err
	}

	return account.QRCodePath, nil
}

