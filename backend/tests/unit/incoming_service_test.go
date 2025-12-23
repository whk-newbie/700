package unit

import (
	"line-management/internal/models"
	"line-management/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IncomingServiceTestSuite 进线处理服务测试套件
type IncomingServiceTestSuite struct {
	suite.Suite
	incomingService *services.IncomingService
	callbackCalled  bool
	callbackData    struct {
		groupID        uint
		lineAccountID  uint
		incomingLineID string
		isDuplicate    bool
	}
}

// SetupSuite 在所有测试开始前执行一次
func (suite *IncomingServiceTestSuite) SetupSuite() {
	// 初始化测试数据库
	SetupTestDB(suite.T())
	
	// 创建带回调的进线服务
	suite.incomingService = services.NewIncomingService(func(groupID uint, lineAccountID uint, incomingLineID string, isDuplicate bool) {
		suite.callbackCalled = true
		suite.callbackData.groupID = groupID
		suite.callbackData.lineAccountID = lineAccountID
		suite.callbackData.incomingLineID = incomingLineID
		suite.callbackData.isDuplicate = isDuplicate
	})
}

// TearDownSuite 在所有测试结束后执行一次
func (suite *IncomingServiceTestSuite) TearDownSuite() {
	TeardownTestDB(suite.T(), TestDB)
}

// SetupTest 在每个测试开始前执行
func (suite *IncomingServiceTestSuite) SetupTest() {
	// 清理测试数据
	CleanupTestData(suite.T(), TestDB)
	// 重置回调标志
	suite.callbackCalled = false
}

// TestProcessIncoming_NoDuplicate 测试处理进线 - 无重复
func (suite *IncomingServiceTestSuite) TestProcessIncoming_NoDuplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 准备进线数据
	data := &services.IncomingData{
		LineAccountID:  account.LineID,
		IncomingLineID: "new_incoming_line_id_001",
		PlatformType:   "line",
		DisplayName:    "Test User",
		AvatarURL:      "https://example.com/avatar.jpg",
		PhoneNumber:    "1234567890",
		Timestamp:      "2025-12-23T10:00:00Z",
	}
	
	// 处理进线
	isDuplicate, err := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "新进线不应该被判定为重复")
	
	// 验证进线日志已创建
	var log models.IncomingLog
	err = TestDB.Where("incoming_line_id = ?", data.IncomingLineID).First(&log).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), account.ID, log.LineAccountID)
	assert.Equal(suite.T(), group.ID, log.GroupID)
	assert.Equal(suite.T(), data.IncomingLineID, log.IncomingLineID)
	assert.Equal(suite.T(), data.DisplayName, log.DisplayName)
	assert.Equal(suite.T(), data.AvatarURL, log.AvatarURL)
	assert.Equal(suite.T(), data.PhoneNumber, log.PhoneNumber)
	assert.False(suite.T(), log.IsDuplicate)
	
	// 验证账号统计已更新
	var accountStats models.LineAccountStats
	err = TestDB.Where("line_account_id = ?", account.ID).First(&accountStats).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, accountStats.TotalIncoming, "总进线数应该为1")
	assert.Equal(suite.T(), 1, accountStats.TodayIncoming, "今日进线数应该为1")
	assert.Equal(suite.T(), 0, accountStats.DuplicateIncoming, "重复进线数应该为0")
	assert.Equal(suite.T(), 0, accountStats.TodayDuplicate, "今日重复数应该为0")
	
	// 验证分组统计已更新
	var groupStats models.GroupStats
	err = TestDB.Where("group_id = ?", group.ID).First(&groupStats).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, groupStats.TotalIncoming, "总进线数应该为1")
	assert.Equal(suite.T(), 1, groupStats.TodayIncoming, "今日进线数应该为1")
	assert.Equal(suite.T(), 0, groupStats.DuplicateIncoming, "重复进线数应该为0")
	assert.Equal(suite.T(), 0, groupStats.TodayDuplicate, "今日重复数应该为0")
	
	// 验证底库已添加
	var contact models.ContactPool
	err = TestDB.Where("line_id = ?", data.IncomingLineID).First(&contact).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), group.ID, contact.GroupID)
	assert.Equal(suite.T(), data.IncomingLineID, contact.LineID)
	assert.Equal(suite.T(), "line", contact.PlatformType)
}

// TestProcessIncoming_Duplicate 测试处理进线 - 有重复
func (suite *IncomingServiceTestSuite) TestProcessIncoming_Duplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 准备进线数据
	incomingLineID := "duplicate_incoming_line_id_002"
	data := &services.IncomingData{
		LineAccountID:  account.LineID,
		IncomingLineID: incomingLineID,
		PlatformType:   "line",
		DisplayName:    "Test User",
	}
	
	// 第一次处理进线（不重复）
	isDuplicate1, err1 := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	assert.NoError(suite.T(), err1)
	assert.False(suite.T(), isDuplicate1, "第一次进线不应该被判定为重复")
	
	// 第二次处理相同的进线（应该重复）
	isDuplicate2, err2 := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	assert.NoError(suite.T(), err2)
	assert.True(suite.T(), isDuplicate2, "第二次进线应该被判定为重复")
	
	// 验证进线日志有两条
	var count int64
	TestDB.Model(&models.IncomingLog{}).Where("incoming_line_id = ?", incomingLineID).Count(&count)
	assert.Equal(suite.T(), int64(2), count, "应该有2条进线日志")
	
	// 验证第二条日志标记为重复
	var logs []models.IncomingLog
	TestDB.Where("incoming_line_id = ?", incomingLineID).Order("id ASC").Find(&logs)
	assert.Len(suite.T(), logs, 2)
	assert.False(suite.T(), logs[0].IsDuplicate, "第一条日志不应该标记为重复")
	assert.True(suite.T(), logs[1].IsDuplicate, "第二条日志应该标记为重复")
	
	// 验证账号统计已更新
	var accountStats models.LineAccountStats
	err := TestDB.Where("line_account_id = ?", account.ID).First(&accountStats).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, accountStats.TotalIncoming, "总进线数应该为2")
	assert.Equal(suite.T(), 2, accountStats.TodayIncoming, "今日进线数应该为2")
	assert.Equal(suite.T(), 1, accountStats.DuplicateIncoming, "重复进线数应该为1")
	assert.Equal(suite.T(), 1, accountStats.TodayDuplicate, "今日重复数应该为1")
	
	// 验证分组统计已更新
	var groupStats models.GroupStats
	err = TestDB.Where("group_id = ?", group.ID).First(&groupStats).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, groupStats.TotalIncoming, "总进线数应该为2")
	assert.Equal(suite.T(), 2, groupStats.TodayIncoming, "今日进线数应该为2")
	assert.Equal(suite.T(), 1, groupStats.DuplicateIncoming, "重复进线数应该为1")
	assert.Equal(suite.T(), 1, groupStats.TodayDuplicate, "今日重复数应该为1")
	
	// 验证底库只有一条记录（重复的不添加）
	var contactCount int64
	TestDB.Model(&models.ContactPool{}).Where("line_id = ?", incomingLineID).Count(&contactCount)
	assert.Equal(suite.T(), int64(1), contactCount, "底库应该只有1条记录")
}

// TestProcessIncoming_CurrentMode 测试处理进线 - current模式
func (suite *IncomingServiceTestSuite) TestProcessIncoming_CurrentMode() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "line1", "line")
	account2 := CreateTestLineAccount(suite.T(), TestDB, group2.ID, "line2", "line")
	
	// 准备进线数据
	incomingLineID := "current_mode_incoming_003"
	data := &services.IncomingData{
		LineAccountID:  account1.LineID,
		IncomingLineID: incomingLineID,
		PlatformType:   "line",
	}
	
	// 在分组1中处理进线
	isDuplicate1, err1 := suite.incomingService.ProcessIncoming(data, account1.ID, group1.ID, "current")
	assert.NoError(suite.T(), err1)
	assert.False(suite.T(), isDuplicate1, "分组1中第一次进线不应该被判定为重复")
	
	// 在分组2中处理相同的进线（current模式下不应该重复）
	data.LineAccountID = account2.LineID
	isDuplicate2, err2 := suite.incomingService.ProcessIncoming(data, account2.ID, group2.ID, "current")
	assert.NoError(suite.T(), err2)
	assert.False(suite.T(), isDuplicate2, "分组2中在current模式下不应该被判定为重复")
	
	// 验证两个分组都有进线日志
	var count1, count2 int64
	TestDB.Model(&models.IncomingLog{}).Where("group_id = ? AND incoming_line_id = ?", group1.ID, incomingLineID).Count(&count1)
	TestDB.Model(&models.IncomingLog{}).Where("group_id = ? AND incoming_line_id = ?", group2.ID, incomingLineID).Count(&count2)
	assert.Equal(suite.T(), int64(1), count1, "分组1应该有1条进线日志")
	assert.Equal(suite.T(), int64(1), count2, "分组2应该有1条进线日志")
}

// TestProcessIncoming_GlobalMode 测试处理进线 - global模式
func (suite *IncomingServiceTestSuite) TestProcessIncoming_GlobalMode() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "line1", "line")
	account2 := CreateTestLineAccount(suite.T(), TestDB, group2.ID, "line2", "line")
	
	// 准备进线数据
	incomingLineID := "global_mode_incoming_004"
	data := &services.IncomingData{
		LineAccountID:  account1.LineID,
		IncomingLineID: incomingLineID,
		PlatformType:   "line",
	}
	
	// 在分组1中处理进线
	isDuplicate1, err1 := suite.incomingService.ProcessIncoming(data, account1.ID, group1.ID, "global")
	assert.NoError(suite.T(), err1)
	assert.False(suite.T(), isDuplicate1, "分组1中第一次进线不应该被判定为重复")
	
	// 在分组2中处理相同的进线（global模式下应该重复）
	data.LineAccountID = account2.LineID
	isDuplicate2, err2 := suite.incomingService.ProcessIncoming(data, account2.ID, group2.ID, "global")
	assert.NoError(suite.T(), err2)
	assert.True(suite.T(), isDuplicate2, "分组2中在global模式下应该被判定为重复")
	
	// 验证分组2的进线日志标记为重复
	var log models.IncomingLog
	err := TestDB.Where("group_id = ? AND incoming_line_id = ?", group2.ID, incomingLineID).First(&log).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), log.IsDuplicate, "分组2的进线日志应该标记为重复")
	assert.Equal(suite.T(), "global", log.DuplicateScope, "去重范围应该是global")
}

// TestProcessIncoming_MultipleAccounts 测试处理进线 - 多个账号
func (suite *IncomingServiceTestSuite) TestProcessIncoming_MultipleAccounts() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account1 := CreateTestLineAccount(suite.T(), TestDB, group.ID, "line1", "line")
	account2 := CreateTestLineAccount(suite.T(), TestDB, group.ID, "line2", "line")
	
	// 准备进线数据
	data1 := &services.IncomingData{
		LineAccountID:  account1.LineID,
		IncomingLineID: "multi_account_incoming_005",
		PlatformType:   "line",
	}
	
	data2 := &services.IncomingData{
		LineAccountID:  account2.LineID,
		IncomingLineID: "multi_account_incoming_006",
		PlatformType:   "line",
	}
	
	// 处理两个账号的进线
	isDuplicate1, err1 := suite.incomingService.ProcessIncoming(data1, account1.ID, group.ID, "current")
	assert.NoError(suite.T(), err1)
	assert.False(suite.T(), isDuplicate1)
	
	isDuplicate2, err2 := suite.incomingService.ProcessIncoming(data2, account2.ID, group.ID, "current")
	assert.NoError(suite.T(), err2)
	assert.False(suite.T(), isDuplicate2)
	
	// 验证分组统计汇总了两个账号的进线
	var groupStats models.GroupStats
	err := TestDB.Where("group_id = ?", group.ID).First(&groupStats).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, groupStats.TotalIncoming, "分组总进线数应该为2")
	
	// 验证各账号统计独立
	var accountStats1, accountStats2 models.LineAccountStats
	TestDB.Where("line_account_id = ?", account1.ID).First(&accountStats1)
	TestDB.Where("line_account_id = ?", account2.ID).First(&accountStats2)
	assert.Equal(suite.T(), 1, accountStats1.TotalIncoming, "账号1总进线数应该为1")
	assert.Equal(suite.T(), 1, accountStats2.TotalIncoming, "账号2总进线数应该为1")
}

// TestProcessIncoming_WithAllFields 测试处理进线 - 包含所有字段
func (suite *IncomingServiceTestSuite) TestProcessIncoming_WithAllFields() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 准备完整的进线数据
	data := &services.IncomingData{
		LineAccountID:  account.LineID,
		IncomingLineID: "full_data_incoming_007",
		PlatformType:   "line",
		DisplayName:    "完整用户名",
		AvatarURL:      "https://example.com/avatar.jpg",
		PhoneNumber:    "+86 138 0000 0000",
		Timestamp:      "2025-12-23T12:00:00Z",
	}
	
	// 处理进线
	isDuplicate, err := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate)
	
	// 验证所有字段都保存了
	var log models.IncomingLog
	err = TestDB.Where("incoming_line_id = ?", data.IncomingLineID).First(&log).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), data.DisplayName, log.DisplayName)
	assert.Equal(suite.T(), data.AvatarURL, log.AvatarURL)
	assert.Equal(suite.T(), data.PhoneNumber, log.PhoneNumber)
	assert.NotNil(suite.T(), log.RawData)
}

// TestProcessIncoming_ContactPoolDuplicate 测试处理进线 - 底库已存在
func (suite *IncomingServiceTestSuite) TestProcessIncoming_ContactPoolDuplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 先在底库中创建记录
	incomingLineID := "pool_duplicate_incoming_008"
	CreateTestContactPool(suite.T(), TestDB, group.ID, incomingLineID, "line")
	
	// 准备进线数据
	data := &services.IncomingData{
		LineAccountID:  account.LineID,
		IncomingLineID: incomingLineID,
		PlatformType:   "line",
	}
	
	// 处理进线（虽然进线日志中不重复，但底库已存在）
	isDuplicate, err := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "进线日志中不重复")
	
	// 验证底库中仍然只有一条记录（不会重复添加）
	var count int64
	TestDB.Model(&models.ContactPool{}).Where("line_id = ?", incomingLineID).Count(&count)
	assert.Equal(suite.T(), int64(1), count, "底库应该只有1条记录")
}

// TestProcessIncoming_Transaction 测试处理进线 - 自动创建统计记录
func (suite *IncomingServiceTestSuite) TestProcessIncoming_Transaction() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 删除账号和分组统计记录（模拟异常情况）
	TestDB.Unscoped().Where("line_account_id = ?", account.ID).Delete(&models.LineAccountStats{})
	TestDB.Unscoped().Where("group_id = ?", group.ID).Delete(&models.GroupStats{})
	
	// 准备进线数据
	data := &services.IncomingData{
		LineAccountID:  account.LineID,
		IncomingLineID: "transaction_test_009",
		PlatformType:   "line",
	}
	
	// 处理进线（应该成功，服务会自动创建统计记录）
	isDuplicate, err := suite.incomingService.ProcessIncoming(data, account.ID, group.ID, "current")
	assert.NoError(suite.T(), err, "即使统计记录不存在，服务也应该自动创建并成功处理")
	assert.False(suite.T(), isDuplicate, "新进线不应该被判定为重复")
	
	// 验证进线日志已创建
	var count int64
	TestDB.Model(&models.IncomingLog{}).Where("incoming_line_id = ?", data.IncomingLineID).Count(&count)
	assert.Equal(suite.T(), int64(1), count, "进线日志应该已创建")
	
	// 验证统计记录已自动创建并更新
	var accountStats models.LineAccountStats
	err = TestDB.Where("line_account_id = ?", account.ID).First(&accountStats).Error
	assert.NoError(suite.T(), err, "账号统计记录应该已自动创建")
	assert.Equal(suite.T(), 1, accountStats.TotalIncoming, "账号总进线数应该为1")
	
	var groupStats models.GroupStats
	err = TestDB.Where("group_id = ?", group.ID).First(&groupStats).Error
	assert.NoError(suite.T(), err, "分组统计记录应该已自动创建")
	assert.Equal(suite.T(), 1, groupStats.TotalIncoming, "分组总进线数应该为1")
}

// TestIncomingServiceTestSuite 运行测试套件
func TestIncomingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IncomingServiceTestSuite))
}

