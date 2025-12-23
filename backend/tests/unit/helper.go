package unit

import (
	"fmt"
	"line-management/internal/config"
	"line-management/internal/models"
	"line-management/pkg/database"
	"line-management/pkg/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestDB 测试数据库实例
var TestDB *gorm.DB

// SetupTestDB 初始化测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	// 初始化配置
	config.GlobalConfig = &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "123456",
			DBName:   "line_management_test", // 使用测试数据库
			SSLMode:  "disable",
			TimeZone: "Asia/Shanghai",
		},
		Log: config.LogConfig{
			Level:      "debug",
			FilePath:   "./logs/test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
	}

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// 连接数据库
	if err := database.InitDB(); err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	TestDB = database.GetDB()

	// 清理测试数据（确保从干净状态开始）
	CleanupTestData(t, TestDB)

	return TestDB
}

// CleanupTestData 清理测试数据
func CleanupTestData(t *testing.T, db *gorm.DB) {
	// 按照外键依赖顺序删除（从子表到父表）
	tables := []interface{}{
		&models.IncomingLog{},
		&models.FollowUpRecord{},
		&models.Customer{},
		&models.ContactPool{},
		&models.ImportBatch{},
		&models.LineAccountStats{},
		&models.LineAccount{},
		&models.GroupStats{},
		&models.Group{},
		&models.User{},
	}

	for _, table := range tables {
		// 硬删除所有记录（包括软删除的）
		if err := db.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			t.Logf("Warning: Failed to clean table %T: %v", table, err)
		}
	}
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB, role string) *models.User {
	// 使用时间戳+随机数确保唯一性
	user := &models.User{
		Username:     fmt.Sprintf("test_user_%s_%d_%d", role, time.Now().UnixNano(), time.Now().Unix()),
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password"的bcrypt hash
		Role:         role,
		IsActive:     true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err, "Failed to create test user")
	return user
}

// CreateTestGroup 创建测试分组
func CreateTestGroup(t *testing.T, db *gorm.DB, userID uint, activationCode string) *models.Group {
	if activationCode == "" {
		activationCode = fmt.Sprintf("TEST%04d", time.Now().UnixNano()%10000)
	}

	accountLimit := 10
	group := &models.Group{
		ActivationCode: activationCode,
		UserID:         userID,
		AccountLimit:   &accountLimit,
		IsActive:       true,
		DedupScope:     "current", // 默认当前分组去重
		ResetTime:      "00:00:00",
		Remark:         fmt.Sprintf("Test Group %d", time.Now().UnixNano()),
	}
	err := db.Create(group).Error
	assert.NoError(t, err, "Failed to create test group")

	// 创建对应的统计记录
	stats := &models.GroupStats{
		GroupID:              group.ID,
		TotalAccounts:        0,
		OnlineAccounts:       0,
		LineAccounts:         0,
		LineBusinessAccounts: 0,
		TotalIncoming:        0,
		TodayIncoming:        0,
		DuplicateIncoming:    0,
		TodayDuplicate:       0,
	}
	err = db.Create(stats).Error
	assert.NoError(t, err, "Failed to create test group stats")

	return group
}

// CreateTestLineAccount 创建测试Line账号
func CreateTestLineAccount(t *testing.T, db *gorm.DB, groupID uint, lineID string, platform string) *models.LineAccount {
	if lineID == "" {
		lineID = fmt.Sprintf("test_line_%d", time.Now().UnixNano())
	}
	if platform == "" {
		platform = "line"
	}

	// 获取分组的激活码
	var group models.Group
	db.Where("id = ?", groupID).First(&group)

	account := &models.LineAccount{
		GroupID:        groupID,
		ActivationCode: group.ActivationCode,
		LineID:         lineID,
		DisplayName:    fmt.Sprintf("Test Account %d", time.Now().UnixNano()),
		PlatformType:   platform,
		OnlineStatus:   "offline",
	}
	err := db.Create(account).Error
	assert.NoError(t, err, "Failed to create test line account")

	// 创建对应的统计记录
	stats := &models.LineAccountStats{
		LineAccountID:     account.ID,
		TotalIncoming:     0,
		TodayIncoming:     0,
		DuplicateIncoming: 0,
		TodayDuplicate:    0,
	}
	err = db.Create(stats).Error
	assert.NoError(t, err, "Failed to create test line account stats")

	return account
}

// CreateTestContactPool 创建测试底库记录
func CreateTestContactPool(t *testing.T, db *gorm.DB, groupID uint, lineID string, platform string) *models.ContactPool {
	// 获取分组的激活码
	var group models.Group
	db.Where("id = ?", groupID).First(&group)

	contact := &models.ContactPool{
		SourceType:     "platform",
		GroupID:        groupID,
		ActivationCode: group.ActivationCode,
		LineID:         lineID,
		PlatformType:   platform,
		DisplayName:    "Test Contact",
	}
	err := db.Create(contact).Error
	assert.NoError(t, err, "Failed to create test contact pool")
	return contact
}

// CreateTestIncomingLog 创建测试进线日志
func CreateTestIncomingLog(t *testing.T, db *gorm.DB, accountID, groupID uint, incomingLineID string, isDuplicate bool, platform string) *models.IncomingLog {
	log := &models.IncomingLog{
		LineAccountID:  accountID,
		GroupID:        groupID,
		IncomingLineID: incomingLineID,
		IsDuplicate:    isDuplicate,
		IncomingTime:   time.Now(),
	}
	err := db.Create(log).Error
	assert.NoError(t, err, "Failed to create test incoming log")
	return log
}

// CreateTestIncomingLogWithTime 创建测试进线日志（指定时间）
func CreateTestIncomingLogWithTime(t *testing.T, db *gorm.DB, accountID, groupID uint, incomingLineID string, isDuplicate bool, platform string, incomingTime time.Time) *models.IncomingLog {
	log := &models.IncomingLog{
		LineAccountID:  accountID,
		GroupID:        groupID,
		IncomingLineID: incomingLineID,
		IsDuplicate:    isDuplicate,
		IncomingTime:   incomingTime,
	}
	err := db.Create(log).Error
	assert.NoError(t, err, "Failed to create test incoming log with time")
	return log
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, db *gorm.DB) {
	CleanupTestData(t, db)
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

