package unit

import (
	"line-management/internal/models"
	"line-management/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DedupServiceTestSuite 去重服务测试套件
type DedupServiceTestSuite struct {
	suite.Suite
	dedupService *services.DedupService
}

// SetupSuite 在所有测试开始前执行一次
func (suite *DedupServiceTestSuite) SetupSuite() {
	// 初始化测试数据库
	SetupTestDB(suite.T())
	suite.dedupService = services.NewDedupService()
}

// TearDownSuite 在所有测试结束后执行一次
func (suite *DedupServiceTestSuite) TearDownSuite() {
	TeardownTestDB(suite.T(), TestDB)
}

// SetupTest 在每个测试开始前执行
func (suite *DedupServiceTestSuite) SetupTest() {
	// 清理测试数据
	CleanupTestData(suite.T(), TestDB)
}

// TestCheckDuplicateCurrent_NoDuplicate 测试当前分组去重 - 无重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateCurrent_NoDuplicate() {
	// 创建测试用户和分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 检查不存在的incoming_line_id
	isDuplicate, err := suite.dedupService.CheckDuplicateCurrent(group.ID, "new_line_id_123")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "新的Line ID不应该被判定为重复")
}

// TestCheckDuplicateCurrent_Duplicate 测试当前分组去重 - 有重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateCurrent_Duplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建一条进线记录
	incomingLineID := "duplicate_line_id_123"
	CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, incomingLineID, false, "line")
	
	// 检查相同的incoming_line_id（应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckDuplicateCurrent(group.ID, incomingLineID)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDuplicate, "已存在的Line ID应该被判定为重复")
}

// TestCheckDuplicateCurrent_DifferentGroup 测试当前分组去重 - 不同分组不算重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateCurrent_DifferentGroup() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "", "line")
	
	// 在分组1中创建进线记录
	incomingLineID := "same_line_id_456"
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, incomingLineID, false, "line")
	
	// 在分组2中检查相同的incoming_line_id（在current模式下不应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckDuplicateCurrent(group2.ID, incomingLineID)
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "其他分组的Line ID在current模式下不应该被判定为重复")
}

// TestCheckDuplicateGlobal_NoDuplicate 测试全局去重 - 无重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateGlobal_NoDuplicate() {
	// 检查不存在的incoming_line_id
	isDuplicate, err := suite.dedupService.CheckDuplicateGlobal("new_global_line_id_789")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "新的Line ID不应该被判定为重复")
}

// TestCheckDuplicateGlobal_Duplicate 测试全局去重 - 有重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateGlobal_Duplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建一条进线记录
	incomingLineID := "global_duplicate_line_id_999"
	CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, incomingLineID, false, "line")
	
	// 检查相同的incoming_line_id（应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckDuplicateGlobal(incomingLineID)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDuplicate, "已存在的Line ID应该被判定为重复")
}

// TestCheckDuplicateGlobal_CrossGroup 测试全局去重 - 跨分组重复
func (suite *DedupServiceTestSuite) TestCheckDuplicateGlobal_CrossGroup() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "", "line")
	
	// 在分组1中创建进线记录
	incomingLineID := "cross_group_line_id_111"
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, incomingLineID, false, "line")
	
	// 全局检查相同的incoming_line_id（应该被判定为重复，即使来自不同分组）
	isDuplicate, err := suite.dedupService.CheckDuplicateGlobal(incomingLineID)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDuplicate, "全局模式下，其他分组的Line ID也应该被判定为重复")
	
	// 确认group2中确实没有这条记录
	var count int64
	TestDB.Model(&models.IncomingLog{}).
		Where("group_id = ? AND incoming_line_id = ?", group2.ID, incomingLineID).
		Count(&count)
	assert.Equal(suite.T(), int64(0), count, "Group2中不应该有这条记录")
}

// TestCheckDuplicate_CurrentMode 测试根据配置检查去重 - current模式
func (suite *DedupServiceTestSuite) TestCheckDuplicate_CurrentMode() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "", "line")
	
	// 在分组1中创建进线记录
	incomingLineID := "current_mode_line_id_222"
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, incomingLineID, false, "line")
	
	// 在分组1中检查（应该重复）
	isDuplicate1, scope1, err1 := suite.dedupService.CheckDuplicate(group1.ID, incomingLineID, "current")
	assert.NoError(suite.T(), err1)
	assert.True(suite.T(), isDuplicate1, "在分组1中应该被判定为重复")
	assert.Equal(suite.T(), "current", scope1)
	
	// 在分组2中检查（不应该重复）
	isDuplicate2, scope2, err2 := suite.dedupService.CheckDuplicate(group2.ID, incomingLineID, "current")
	assert.NoError(suite.T(), err2)
	assert.False(suite.T(), isDuplicate2, "在分组2中不应该被判定为重复")
	assert.Equal(suite.T(), "current", scope2)
}

// TestCheckDuplicate_GlobalMode 测试根据配置检查去重 - global模式
func (suite *DedupServiceTestSuite) TestCheckDuplicate_GlobalMode() {
	// 创建两个不同的分组
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "", "line")
	
	// 在分组1中创建进线记录
	incomingLineID := "global_mode_line_id_333"
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, incomingLineID, false, "line")
	
	// 在分组1中检查（应该重复）
	isDuplicate1, scope1, err1 := suite.dedupService.CheckDuplicate(group1.ID, incomingLineID, "global")
	assert.NoError(suite.T(), err1)
	assert.True(suite.T(), isDuplicate1, "在分组1中应该被判定为重复")
	assert.Equal(suite.T(), "global", scope1)
	
	// 在分组2中检查（全局模式下也应该重复）
	isDuplicate2, scope2, err2 := suite.dedupService.CheckDuplicate(group2.ID, incomingLineID, "global")
	assert.NoError(suite.T(), err2)
	assert.True(suite.T(), isDuplicate2, "全局模式下，分组2中也应该被判定为重复")
	assert.Equal(suite.T(), "global", scope2)
}

// TestCheckContactPoolDuplicate_NoDuplicate 测试底库去重 - 无重复
func (suite *DedupServiceTestSuite) TestCheckContactPoolDuplicate_NoDuplicate() {
	// 检查不存在的line_id
	isDuplicate, err := suite.dedupService.CheckContactPoolDuplicate("new_pool_line_id_444", "line")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "底库中不存在的Line ID不应该被判定为重复")
}

// TestCheckContactPoolDuplicate_Duplicate 测试底库去重 - 有重复
func (suite *DedupServiceTestSuite) TestCheckContactPoolDuplicate_Duplicate() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 在底库中创建记录
	lineID := "pool_duplicate_line_id_555"
	CreateTestContactPool(suite.T(), TestDB, group.ID, lineID, "line")
	
	// 检查相同的line_id（应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckContactPoolDuplicate(lineID, "line")
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDuplicate, "底库中已存在的Line ID应该被判定为重复")
}

// TestCheckContactPoolDuplicate_DifferentPlatform 测试底库去重 - 不同平台不算重复
func (suite *DedupServiceTestSuite) TestCheckContactPoolDuplicate_DifferentPlatform() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 在底库中创建line平台的记录
	lineID := "pool_platform_line_id_666"
	CreateTestContactPool(suite.T(), TestDB, group.ID, lineID, "line")
	
	// 检查line_business平台的相同line_id（不应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckContactPoolDuplicate(lineID, "line_business")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "不同平台的Line ID不应该被判定为重复")
}

// TestCheckContactPoolDuplicate_DeletedRecord 测试底库去重 - 已删除的记录不算重复
func (suite *DedupServiceTestSuite) TestCheckContactPoolDuplicate_DeletedRecord() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 在底库中创建记录
	lineID := "pool_deleted_line_id_777"
	contact := CreateTestContactPool(suite.T(), TestDB, group.ID, lineID, "line")
	
	// 软删除该记录
	err := TestDB.Delete(contact).Error
	assert.NoError(suite.T(), err)
	
	// 检查相同的line_id（已删除的记录不应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckContactPoolDuplicate(lineID, "line")
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDuplicate, "已删除的底库记录不应该被判定为重复")
}

// TestCheckDuplicate_MultipleRecords 测试去重逻辑 - 多条相同记录
func (suite *DedupServiceTestSuite) TestCheckDuplicate_MultipleRecords() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建多条相同incoming_line_id的进线记录（模拟多次重复进线）
	incomingLineID := "multiple_records_line_id_888"
	CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, incomingLineID, false, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, incomingLineID, true, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, incomingLineID, true, "line")
	
	// 检查去重（即使有多条记录，也应该被判定为重复）
	isDuplicate, err := suite.dedupService.CheckDuplicateCurrent(group.ID, incomingLineID)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDuplicate, "多条相同Line ID的记录应该被判定为重复")
}

// TestDedupServiceTestSuite 运行测试套件
func TestDedupServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DedupServiceTestSuite))
}

