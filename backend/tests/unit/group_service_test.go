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

// GroupServiceTestSuite 分组服务测试套件
type GroupServiceTestSuite struct {
	suite.Suite
	db            *gorm.DB
	service       *services.GroupService
	adminUser     *models.User
	normalUser    *models.User
	testGroup     *models.Group
}

// SetupSuite 测试套件初始化
func (suite *GroupServiceTestSuite) SetupSuite() {
	suite.db = SetupTestDB(suite.T())
	suite.service = services.NewGroupService()
	
	// 创建测试用户
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.normalUser = CreateTestUser(suite.T(), suite.db, "user")
	
	// 创建测试分组
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
}

// TearDownTest 每个测试后清理
func (suite *GroupServiceTestSuite) TearDownTest() {
	CleanupTestData(suite.T(), suite.db)
	// 重新创建基础数据
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.normalUser = CreateTestUser(suite.T(), suite.db, "user")
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
}

// TestGenerateActivationCode 测试生成激活码
func (suite *GroupServiceTestSuite) TestGenerateActivationCode() {
	code, err := suite.service.GenerateActivationCode()
	suite.NoError(err)
	suite.NotEmpty(code)
	suite.Len(code, 8)
	
	// 验证激活码格式（大写字母+数字）
	for _, char := range code {
		suite.True(
			(char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'),
			"激活码应只包含大写字母和数字",
		)
	}
}

// TestGenerateActivationCode_Uniqueness 测试激活码唯一性
func (suite *GroupServiceTestSuite) TestGenerateActivationCode_Uniqueness() {
	codes := make(map[string]bool)
	for i := 0; i < 10; i++ {
		code, err := suite.service.GenerateActivationCode()
		suite.NoError(err)
		suite.False(codes[code], "激活码应该唯一")
		codes[code] = true
	}
}

// TestCreateGroup_Success 测试成功创建分组
func (suite *GroupServiceTestSuite) TestCreateGroup_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateGroupRequest{
		UserID:       suite.adminUser.ID,
		AccountLimit: intPtr(10),
		IsActive:     true,
		Remark:       "测试分组",
		Description:  "这是一个测试分组",
		Category:     "test",
		DedupScope:   "current",
		ResetTime:    "09:00:00",
	}
	
	group, err := suite.service.CreateGroup(c, req)
	suite.NoError(err)
	suite.NotNil(group)
	suite.NotEmpty(group.ActivationCode)
	suite.Equal(suite.adminUser.ID, group.UserID)
	suite.Equal("test", group.Category)
	suite.Equal("current", group.DedupScope)
	suite.Equal("09:00:00", group.ResetTime)
	
	// 验证统计记录已创建
	var stats models.GroupStats
	err = suite.db.Where("group_id = ?", group.ID).First(&stats).Error
	suite.NoError(err)
	suite.Equal(group.ID, stats.GroupID)
}

// TestCreateGroup_UserNotExists 测试用户不存在
func (suite *GroupServiceTestSuite) TestCreateGroup_UserNotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateGroupRequest{
		UserID:       99999, // 不存在的用户ID
		AccountLimit: intPtr(10),
		IsActive:     true,
	}
	
	group, err := suite.service.CreateGroup(c, req)
	suite.Error(err)
	suite.Nil(group)
	suite.Contains(err.Error(), "用户不存在")
}

// TestCreateGroup_UserInactive 测试用户被禁用
func (suite *GroupServiceTestSuite) TestCreateGroup_UserInactive() {
	// 创建被禁用的用户
	inactiveUser := CreateTestUser(suite.T(), suite.db, "user")
	inactiveUser.IsActive = false
	suite.db.Save(inactiveUser)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateGroupRequest{
		UserID:       inactiveUser.ID,
		AccountLimit: intPtr(10),
		IsActive:     true,
	}
	
	group, err := suite.service.CreateGroup(c, req)
	suite.Error(err)
	suite.Nil(group)
	suite.Contains(err.Error(), "用户已被禁用")
}

// TestCreateGroup_WithPassword 测试创建带密码的分组
func (suite *GroupServiceTestSuite) TestCreateGroup_WithPassword() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateGroupRequest{
		UserID:       suite.adminUser.ID,
		AccountLimit: intPtr(10),
		IsActive:     true,
		LoginPassword: "password123",
	}
	
	group, err := suite.service.CreateGroup(c, req)
	suite.NoError(err)
	suite.NotNil(group)
	suite.NotEmpty(group.LoginPassword)
	suite.NotEqual("password123", group.LoginPassword) // 应该是加密后的
}

// TestCreateGroup_DefaultValues 测试默认值
func (suite *GroupServiceTestSuite) TestCreateGroup_DefaultValues() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateGroupRequest{
		UserID:       suite.adminUser.ID,
		AccountLimit: intPtr(10),
		IsActive:     true,
		// 不提供Category、DedupScope、ResetTime，应该使用默认值
	}
	
	group, err := suite.service.CreateGroup(c, req)
	suite.NoError(err)
	suite.NotNil(group)
	suite.Equal("default", group.Category)
	suite.Equal("current", group.DedupScope)
	suite.Equal("09:00:00", group.ResetTime)
}

// TestGetGroupList_Success 测试获取分组列表
func (suite *GroupServiceTestSuite) TestGetGroupList_Success() {
	// 创建多个测试分组
	CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0002")
	CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0003")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.GroupQueryParams{
		Page:     1,
		PageSize: 10,
	}
	
	list, total, err := suite.service.GetGroupList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(3)) // 至少包含suite.testGroup, group1, group2
	suite.GreaterOrEqual(len(list), 3)
}

// TestGetGroupList_WithFilter 测试带筛选的列表查询
func (suite *GroupServiceTestSuite) TestGetGroupList_WithFilter() {
	// 创建不同状态的分组
	activeGroup := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0004")
	inactiveGroup := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0005")
	inactiveGroup.IsActive = false
	suite.db.Save(inactiveGroup)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	// 筛选激活的分组
	isActive := true
	params := &schemas.GroupQueryParams{
		Page:     1,
		PageSize: 10,
		IsActive: &isActive,
	}
	
	list, total, err := suite.service.GetGroupList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	for _, g := range list {
		suite.True(g.IsActive, "应该只返回激活的分组")
	}
	
	// 验证包含activeGroup
	found := false
	for _, g := range list {
		if g.ID == activeGroup.ID {
			found = true
			break
		}
	}
	suite.True(found, "应该包含activeGroup")
}

// TestGetGroupList_WithSearch 测试搜索功能
func (suite *GroupServiceTestSuite) TestGetGroupList_WithSearch() {
	// 创建带特定备注的分组
	group := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0006")
	group.Remark = "特殊备注分组"
	suite.db.Save(group)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.GroupQueryParams{
		Page:     1,
		PageSize: 10,
		Search:   "特殊备注",
	}
	
	list, total, err := suite.service.GetGroupList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	
	found := false
	for _, g := range list {
		if g.ID == group.ID {
			found = true
			break
		}
	}
	suite.True(found, "应该找到包含'特殊备注'的分组")
}

// TestGetGroupByID_Success 测试根据ID获取分组
func (suite *GroupServiceTestSuite) TestGetGroupByID_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	group, err := suite.service.GetGroupByID(c, suite.testGroup.ID)
	suite.NoError(err)
	suite.NotNil(group)
	suite.Equal(suite.testGroup.ID, group.ID)
	suite.Equal(suite.testGroup.ActivationCode, group.ActivationCode)
}

// TestGetGroupByID_NotExists 测试分组不存在
func (suite *GroupServiceTestSuite) TestGetGroupByID_NotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	group, err := suite.service.GetGroupByID(c, 99999)
	suite.Error(err)
	suite.Nil(group)
	suite.Contains(err.Error(), "分组不存在")
}

// TestUpdateGroup_Success 测试更新分组
func (suite *GroupServiceTestSuite) TestUpdateGroup_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	newLimit := 20
	newActive := false
	req := &schemas.UpdateGroupRequest{
		AccountLimit: &newLimit,
		IsActive:     &newActive,
		Remark:       "更新后的备注",
		Description:  "更新后的描述",
		Category:     "updated",
		DedupScope:   "global",
		ResetTime:    "10:00:00",
	}
	
	group, err := suite.service.UpdateGroup(c, suite.testGroup.ID, req)
	suite.NoError(err)
	suite.NotNil(group)
	suite.Equal(20, *group.AccountLimit)
	suite.False(group.IsActive)
	suite.Equal("更新后的备注", group.Remark)
	suite.Equal("更新后的描述", group.Description)
	suite.Equal("updated", group.Category)
	suite.Equal("global", group.DedupScope)
	suite.Equal("10:00:00", group.ResetTime)
}

// TestUpdateGroup_UpdatePassword 测试更新密码
func (suite *GroupServiceTestSuite) TestUpdateGroup_UpdatePassword() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.UpdateGroupRequest{
		LoginPassword: "newpassword123",
	}
	
	group, err := suite.service.UpdateGroup(c, suite.testGroup.ID, req)
	suite.NoError(err)
	suite.NotNil(group)
	suite.NotEmpty(group.LoginPassword)
	suite.NotEqual("newpassword123", group.LoginPassword) // 应该是加密后的
}

// TestDeleteGroup_Success 测试删除分组（软删除）
func (suite *GroupServiceTestSuite) TestDeleteGroup_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	err := suite.service.DeleteGroup(c, suite.testGroup.ID)
	suite.NoError(err)
	
	// 验证软删除（应该查询不到）
	var group models.Group
	err = suite.db.Where("id = ? AND deleted_at IS NULL", suite.testGroup.ID).First(&group).Error
	suite.Error(err)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound))
	
	// 验证硬删除查询（应该能查到）
	err = suite.db.Unscoped().Where("id = ?", suite.testGroup.ID).First(&group).Error
	suite.NoError(err)
	suite.NotNil(group.DeletedAt)
}

// TestRegenerateActivationCode_Success 测试重新生成激活码
func (suite *GroupServiceTestSuite) TestRegenerateActivationCode_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	oldCode := suite.testGroup.ActivationCode
	newCode, err := suite.service.RegenerateActivationCode(c, suite.testGroup.ID)
	suite.NoError(err)
	suite.NotEmpty(newCode)
	suite.NotEqual(oldCode, newCode)
	suite.Len(newCode, 8)
	
	// 验证数据库中的激活码已更新
	var group models.Group
	suite.db.Where("id = ?", suite.testGroup.ID).First(&group)
	suite.Equal(newCode, group.ActivationCode)
}

// TestGetCategories_Success 测试获取分类列表
func (suite *GroupServiceTestSuite) TestGetCategories_Success() {
	// 创建不同分类的分组
	CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0007")
	group2 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0008")
	group2.Category = "category1"
	suite.db.Save(group2)
	
	group3 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0009")
	group3.Category = "category2"
	suite.db.Save(group3)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	categories, err := suite.service.GetCategories(c)
	suite.NoError(err)
	suite.GreaterOrEqual(len(categories), 2)
	
	// 验证包含创建的分类
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}
	suite.True(categoryMap["category1"] || categoryMap["category2"], "应该包含创建的分类")
}

// TestBatchDeleteGroups_Success 测试批量删除分组
func (suite *GroupServiceTestSuite) TestBatchDeleteGroups_Success() {
	// 创建多个分组
	group1 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0010")
	group2 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0011")
	group3 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0012")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	ids := []uint{group1.ID, group2.ID, group3.ID, 99999} // 包含一个不存在的ID
	successCount, failedIDs, err := suite.service.BatchDeleteGroups(c, ids)
	suite.NoError(err)
	suite.Equal(3, successCount)
	suite.Equal(1, len(failedIDs))
	suite.Equal(uint(99999), failedIDs[0])
	
	// 验证分组已被软删除
	var count int64
	suite.db.Model(&models.Group{}).Where("id IN ? AND deleted_at IS NULL", []uint{group1.ID, group2.ID, group3.ID}).Count(&count)
	suite.Equal(int64(0), count)
}

// TestBatchUpdateGroups_Success 测试批量更新分组
func (suite *GroupServiceTestSuite) TestBatchUpdateGroups_Success() {
	// 创建多个分组
	group1 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0013")
	group2 := CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0014")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	newActive := false
	req := &schemas.BatchUpdateGroupsRequest{
		IsActive:   &newActive,
		Category:   "batch_updated",
		DedupScope: "global",
	}
	
	ids := []uint{group1.ID, group2.ID, 99999} // 包含一个不存在的ID
	successCount, failedIDs, err := suite.service.BatchUpdateGroups(c, ids, req)
	suite.NoError(err)
	suite.Equal(2, successCount)
	suite.Equal(1, len(failedIDs))
	
	// 验证分组已更新
	var updatedGroup1, updatedGroup2 models.Group
	suite.db.Where("id = ?", group1.ID).First(&updatedGroup1)
	suite.db.Where("id = ?", group2.ID).First(&updatedGroup2)
	suite.False(updatedGroup1.IsActive)
	suite.False(updatedGroup2.IsActive)
	suite.Equal("batch_updated", updatedGroup1.Category)
	suite.Equal("batch_updated", updatedGroup2.Category)
	suite.Equal("global", updatedGroup1.DedupScope)
	suite.Equal("global", updatedGroup2.DedupScope)
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

// TestGroupServiceTestSuite 运行测试套件
func TestGroupServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GroupServiceTestSuite))
}

