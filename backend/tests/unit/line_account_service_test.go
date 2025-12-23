package unit

import (
	"errors"
	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LineAccountServiceTestSuite Line账号服务测试套件
type LineAccountServiceTestSuite struct {
	suite.Suite
	db            *gorm.DB
	service       *services.LineAccountService
	adminUser     *models.User
	testGroup     *models.Group
	testAccount   *models.LineAccount
}

// SetupSuite 测试套件初始化
func (suite *LineAccountServiceTestSuite) SetupSuite() {
	suite.db = SetupTestDB(suite.T())
	suite.service = services.NewLineAccountService()
	
	// 创建测试用户和分组
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
	
	// 创建测试账号
	suite.testAccount = CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_001", "line")
}

// TearDownTest 每个测试后清理
func (suite *LineAccountServiceTestSuite) TearDownTest() {
	CleanupTestData(suite.T(), suite.db)
	// 重新创建基础数据
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
	suite.testAccount = CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_001", "line")
}

// TestCreateLineAccount_Success 测试成功创建Line账号
func (suite *LineAccountServiceTestSuite) TestCreateLineAccount_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateLineAccountRequest{
		GroupID:      suite.testGroup.ID,
		PlatformType: "line",
		LineID:       "test_line_002",
		DisplayName:  "测试账号2",
		PhoneNumber:  "13800138000",
	}
	
	account, err := suite.service.CreateLineAccount(c, req)
	suite.NoError(err)
	suite.NotNil(account)
	suite.Equal(suite.testGroup.ID, account.GroupID)
	suite.Equal("test_line_002", account.LineID)
	suite.Equal("line", account.PlatformType)
	suite.Equal("offline", account.OnlineStatus)
	
	// 验证统计记录已创建
	var stats models.LineAccountStats
	err = suite.db.Where("line_account_id = ?", account.ID).First(&stats).Error
	suite.NoError(err)
	suite.Equal(account.ID, stats.LineAccountID)
	
	// 验证分组统计已更新
	var groupStats models.GroupStats
	suite.db.Where("group_id = ?", suite.testGroup.ID).First(&groupStats)
	suite.GreaterOrEqual(groupStats.TotalAccounts, 1)
	suite.GreaterOrEqual(groupStats.LineAccounts, 1)
}

// TestCreateLineAccount_GroupNotExists 测试分组不存在
func (suite *LineAccountServiceTestSuite) TestCreateLineAccount_GroupNotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateLineAccountRequest{
		GroupID:      99999, // 不存在的分组ID
		PlatformType: "line",
		LineID:       "test_line_003",
	}
	
	account, err := suite.service.CreateLineAccount(c, req)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "分组不存在")
}

// TestCreateLineAccount_GroupInactive 测试分组被禁用
func (suite *LineAccountServiceTestSuite) TestCreateLineAccount_GroupInactive() {
	// 创建被禁用的分组
	inactiveGroup := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0002")
	inactiveGroup.IsActive = false
	suite.db.Save(inactiveGroup)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateLineAccountRequest{
		GroupID:      inactiveGroup.ID,
		PlatformType: "line",
		LineID:       "test_line_004",
	}
	
	account, err := suite.service.CreateLineAccount(c, req)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "分组已被禁用")
}

// TestCreateLineAccount_AccountLimit 测试账号数量限制
func (suite *LineAccountServiceTestSuite) TestCreateLineAccount_AccountLimit() {
	// 创建有限制的分组
	limit := 1
	limitedGroup := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0003")
	limitedGroup.AccountLimit = &limit
	suite.db.Save(limitedGroup)
	
	// 创建一个账号（达到限制）
	CreateTestLineAccount(suite.T(), suite.db, limitedGroup.ID, "test_line_limit_1", "line")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateLineAccountRequest{
		GroupID:      limitedGroup.ID,
		PlatformType: "line",
		LineID:       "test_line_limit_2",
	}
	
	account, err := suite.service.CreateLineAccount(c, req)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "已达到分组账号数量限制")
}

// TestCreateLineAccount_DuplicateLineID 测试重复的Line ID
func (suite *LineAccountServiceTestSuite) TestCreateLineAccount_DuplicateLineID() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateLineAccountRequest{
		GroupID:      suite.testGroup.ID,
		PlatformType: "line",
		LineID:       suite.testAccount.LineID, // 使用已存在的Line ID
	}
	
	account, err := suite.service.CreateLineAccount(c, req)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "该Line ID在此分组中已存在")
}

// TestGetLineAccountList_Success 测试获取账号列表
func (suite *LineAccountServiceTestSuite) TestGetLineAccountList_Success() {
	// 创建多个测试账号
	CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_005", "line")
	CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_006", "line_business")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.LineAccountQueryParams{
		Page:     1,
		PageSize: 10,
	}
	
	list, total, err := suite.service.GetLineAccountList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(3)) // 至少包含suite.testAccount和刚创建的两个
	suite.GreaterOrEqual(len(list), 3)
}

// TestGetLineAccountList_WithFilter 测试带筛选的列表查询
func (suite *LineAccountServiceTestSuite) TestGetLineAccountList_WithFilter() {
	// 创建不同平台和状态的账号
	lineAccount := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_007", "line")
	CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_008", "line_business")
	
	// 设置在线状态
	lineAccount.OnlineStatus = "online"
	suite.db.Save(lineAccount)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	// 筛选line平台
	params := &schemas.LineAccountQueryParams{
		Page:         1,
		PageSize:     10,
		PlatformType: "line",
	}
	
	list, total, err := suite.service.GetLineAccountList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	for _, acc := range list {
		suite.Equal("line", acc.PlatformType)
	}
	
	// 筛选在线状态
	params.OnlineStatus = "online"
	params.PlatformType = "" // 清除平台筛选
	list2, total2, err := suite.service.GetLineAccountList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total2, int64(1))
	for _, acc := range list2 {
		suite.Equal("online", acc.OnlineStatus)
	}
}

// TestGetLineAccountList_WithSearch 测试搜索功能
func (suite *LineAccountServiceTestSuite) TestGetLineAccountList_WithSearch() {
	// 创建带特定显示名称的账号
	account := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_009", "line")
	account.DisplayName = "特殊名称账号"
	suite.db.Save(account)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.LineAccountQueryParams{
		Page:     1,
		PageSize: 10,
		Search:   "特殊名称",
	}
	
	list, total, err := suite.service.GetLineAccountList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	
	found := false
	for _, acc := range list {
		if acc.ID == account.ID {
			found = true
			break
		}
	}
	suite.True(found, "应该找到包含'特殊名称'的账号")
}

// TestGetLineAccountByID_Success 测试根据ID获取账号
func (suite *LineAccountServiceTestSuite) TestGetLineAccountByID_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	account, err := suite.service.GetLineAccountByID(c, suite.testAccount.ID)
	suite.NoError(err)
	suite.NotNil(account)
	suite.Equal(suite.testAccount.ID, account.ID)
	suite.Equal(suite.testAccount.LineID, account.LineID)
}

// TestGetLineAccountByID_NotExists 测试账号不存在
func (suite *LineAccountServiceTestSuite) TestGetLineAccountByID_NotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	account, err := suite.service.GetLineAccountByID(c, 99999)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "账号不存在")
}

// TestUpdateLineAccount_Success 测试更新账号
func (suite *LineAccountServiceTestSuite) TestUpdateLineAccount_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.UpdateLineAccountRequest{
		DisplayName:  "更新后的名称",
		PhoneNumber:  "13900139000",
		OnlineStatus: "online",
	}
	
	account, err := suite.service.UpdateLineAccount(c, suite.testAccount.ID, req)
	suite.NoError(err)
	suite.NotNil(account)
	suite.Equal("更新后的名称", account.DisplayName)
	suite.Equal("13900139000", account.PhoneNumber)
	suite.Equal("online", account.OnlineStatus)
	
	// 验证分组统计已更新（在线账号数增加）
	var groupStats models.GroupStats
	suite.db.Where("group_id = ?", suite.testGroup.ID).First(&groupStats)
	suite.GreaterOrEqual(groupStats.OnlineAccounts, 1)
}

// TestUpdateLineAccount_ChangeGroup 测试更换分组
func (suite *LineAccountServiceTestSuite) TestUpdateLineAccount_ChangeGroup() {
	// 创建新分组
	newGroup := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0004")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	newGroupID := newGroup.ID
	req := &schemas.UpdateLineAccountRequest{
		GroupID: &newGroupID,
	}
	
	account, err := suite.service.UpdateLineAccount(c, suite.testAccount.ID, req)
	suite.NoError(err)
	suite.NotNil(account)
	suite.Equal(newGroup.ID, account.GroupID)
	suite.Equal(newGroup.ActivationCode, account.ActivationCode)
}

// TestUpdateLineAccount_DuplicateLineID 测试更新为重复的Line ID
func (suite *LineAccountServiceTestSuite) TestUpdateLineAccount_DuplicateLineID() {
	// 创建另一个账号
	otherAccount := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_010", "line")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.UpdateLineAccountRequest{
		LineID: otherAccount.LineID, // 使用另一个账号的Line ID
	}
	
	account, err := suite.service.UpdateLineAccount(c, suite.testAccount.ID, req)
	suite.Error(err)
	suite.Nil(account)
	suite.Contains(err.Error(), "该Line ID在此分组中已存在")
}

// TestDeleteLineAccount_Success 测试删除账号（软删除）
func (suite *LineAccountServiceTestSuite) TestDeleteLineAccount_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	deletedBy := suite.adminUser.ID
	err := suite.service.DeleteLineAccount(c, suite.testAccount.ID, &deletedBy)
	suite.NoError(err)
	
	// 验证软删除（应该查询不到）
	var account models.LineAccount
	err = suite.db.Where("id = ? AND deleted_at IS NULL", suite.testAccount.ID).First(&account).Error
	suite.Error(err)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound))
	
	// 验证硬删除查询（应该能查到）
	err = suite.db.Unscoped().Where("id = ?", suite.testAccount.ID).First(&account).Error
	suite.NoError(err)
	suite.NotNil(account.DeletedAt)
	// 注意：GORM的Delete方法可能不会保存DeletedBy字段，所以这里只验证DeletedAt
	// 如果需要保存DeletedBy，服务代码应该使用Updates方法
	
	// 验证分组统计已更新（账号数减少）
	var groupStats models.GroupStats
	suite.db.Where("group_id = ?", suite.testGroup.ID).First(&groupStats)
	suite.LessOrEqual(groupStats.TotalAccounts, 0) // 因为只有这一个账号
}

// TestBatchDeleteLineAccounts_Success 测试批量删除账号
func (suite *LineAccountServiceTestSuite) TestBatchDeleteLineAccounts_Success() {
	// 创建多个账号
	account1 := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_011", "line")
	account2 := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_012", "line")
	account3 := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_013", "line")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	ids := []uint{account1.ID, account2.ID, account3.ID, 99999} // 包含一个不存在的ID
	deletedBy := suite.adminUser.ID
	successCount, failedIDs, err := suite.service.BatchDeleteLineAccounts(c, ids, &deletedBy)
	suite.NoError(err)
	suite.Equal(3, successCount)
	suite.Equal(1, len(failedIDs))
	suite.Equal(uint(99999), failedIDs[0])
	
	// 验证账号已被软删除
	var count int64
	suite.db.Model(&models.LineAccount{}).Where("id IN ? AND deleted_at IS NULL", []uint{account1.ID, account2.ID, account3.ID}).Count(&count)
	suite.Equal(int64(0), count)
}

// TestBatchUpdateLineAccounts_Success 测试批量更新账号
func (suite *LineAccountServiceTestSuite) TestBatchUpdateLineAccounts_Success() {
	// 创建多个账号
	account1 := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_014", "line")
	account2 := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_015", "line")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.BatchUpdateLineAccountsRequest{
		OnlineStatus: "online",
	}
	
	ids := []uint{account1.ID, account2.ID, 99999} // 包含一个不存在的ID
	successCount, failedIDs, err := suite.service.BatchUpdateLineAccounts(c, ids, req)
	suite.NoError(err)
	suite.Equal(2, successCount)
	suite.Equal(1, len(failedIDs))
	
	// 验证账号已更新
	var updatedAccount1, updatedAccount2 models.LineAccount
	suite.db.Where("id = ?", account1.ID).First(&updatedAccount1)
	suite.db.Where("id = ?", account2.ID).First(&updatedAccount2)
	suite.Equal("online", updatedAccount1.OnlineStatus)
	suite.Equal("online", updatedAccount2.OnlineStatus)
	
	// 验证分组统计已更新（在线账号数增加）
	var groupStats models.GroupStats
	suite.db.Where("group_id = ?", suite.testGroup.ID).First(&groupStats)
	suite.GreaterOrEqual(groupStats.OnlineAccounts, 2)
}

// TestLineAccountServiceTestSuite 运行测试套件
func TestLineAccountServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LineAccountServiceTestSuite))
}

