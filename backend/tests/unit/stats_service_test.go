package unit

import (
	"line-management/internal/models"
	"line-management/internal/services"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// StatsServiceTestSuite 统计服务测试套件
type StatsServiceTestSuite struct {
	suite.Suite
	statsService *services.StatsService
}

// SetupSuite 在所有测试开始前执行一次
func (suite *StatsServiceTestSuite) SetupSuite() {
	// 初始化测试数据库
	SetupTestDB(suite.T())
	suite.statsService = services.NewStatsService()
}

// TearDownSuite 在所有测试结束后执行一次
func (suite *StatsServiceTestSuite) TearDownSuite() {
	TeardownTestDB(suite.T(), TestDB)
}

// SetupTest 在每个测试开始前执行
func (suite *StatsServiceTestSuite) SetupTest() {
	// 清理测试数据
	CleanupTestData(suite.T(), TestDB)
}

// createTestContext 创建测试用的gin context（模拟管理员权限）
func (suite *StatsServiceTestSuite) createTestContext() *gin.Context {
	c, _ := gin.CreateTestContext(nil)
	// 设置管理员权限，不应用数据过滤
	c.Set("role", "admin")
	c.Set("data_filter", nil)
	return c
}

// createUserContext 创建测试用的普通用户gin context
func (suite *StatsServiceTestSuite) createUserContext(userID uint) *gin.Context {
	c, _ := gin.CreateTestContext(nil)
	c.Set("role", "user")
	c.Set("user_id", userID)
	c.Set("data_filter", map[string]interface{}{
		"user_id": userID,
	})
	return c
}

// createSubAccountContext 创建测试用的子账号gin context
func (suite *StatsServiceTestSuite) createSubAccountContext(groupID uint) *gin.Context {
	c, _ := gin.CreateTestContext(nil)
	c.Set("role", "subaccount")
	c.Set("group_id", groupID)
	c.Set("data_filter", map[string]interface{}{
		"group_id": groupID,
	})
	return c
}

// TestGetGroupStats_Exists 测试获取分组统计 - 存在的分组
func (suite *StatsServiceTestSuite) TestGetGroupStats_Exists() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 获取分组统计
	stats, err := suite.statsService.GetGroupStats(group.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), group.ID, stats.GroupID)
	assert.Equal(suite.T(), 0, stats.TotalIncoming, "初始进线数应该为0")
	assert.Equal(suite.T(), 0, stats.TodayIncoming, "初始今日进线数应该为0")
	assert.Equal(suite.T(), 0, stats.DuplicateIncoming, "初始重复进线数应该为0")
	assert.Equal(suite.T(), 0, stats.TodayDuplicate, "初始今日重复数应该为0")
}

// TestGetGroupStats_NotExists 测试获取分组统计 - 不存在的分组（自动创建）
func (suite *StatsServiceTestSuite) TestGetGroupStats_NotExists() {
	// 创建测试数据（只创建分组，不创建统计记录）
	user := CreateTestUser(suite.T(), TestDB, "admin")
	accountLimit := 10
	group := &models.Group{
		Remark:         "Test Group Without Stats",
		ActivationCode: "NOSTATS",
		IsActive:       true,
		UserID:         user.ID,
		AccountLimit:   &accountLimit,
		DedupScope:     "current",
		ResetTime:      "00:00:00",
	}
	err := TestDB.Create(group).Error
	assert.NoError(suite.T(), err)
	
	// 删除自动创建的统计记录
	TestDB.Unscoped().Where("group_id = ?", group.ID).Delete(&models.GroupStats{})
	
	// 获取分组统计（应该自动创建）
	stats, err := suite.statsService.GetGroupStats(group.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), group.ID, stats.GroupID)
}

// TestGetAccountStats_Exists 测试获取账号统计 - 存在的账号
func (suite *StatsServiceTestSuite) TestGetAccountStats_Exists() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 获取账号统计
	stats, err := suite.statsService.GetAccountStats(account.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), account.ID, stats.LineAccountID)
	assert.Equal(suite.T(), 0, stats.TotalIncoming, "初始进线数应该为0")
	assert.Equal(suite.T(), 0, stats.TodayIncoming, "初始今日进线数应该为0")
	assert.Equal(suite.T(), 0, stats.DuplicateIncoming, "初始重复进线数应该为0")
	assert.Equal(suite.T(), 0, stats.TodayDuplicate, "初始今日重复数应该为0")
}

// TestGetAccountStats_NotExists 测试获取账号统计 - 不存在的账号（自动创建）
func (suite *StatsServiceTestSuite) TestGetAccountStats_NotExists() {
	// 创建测试数据（只创建账号，不创建统计记录）
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := &models.LineAccount{
		GroupID:        group.ID,
		ActivationCode: group.ActivationCode,
		LineID:         "test_line_no_stats",
		DisplayName:    "Test Account Without Stats",
		PlatformType:   "line",
		OnlineStatus:   "offline",
	}
	err := TestDB.Create(account).Error
	assert.NoError(suite.T(), err)
	
	// 删除自动创建的统计记录
	TestDB.Unscoped().Where("line_account_id = ?", account.ID).Delete(&models.LineAccountStats{})
	
	// 获取账号统计（应该自动创建）
	stats, err := suite.statsService.GetAccountStats(account.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), account.ID, stats.LineAccountID)
}

// TestGetOverviewStats 测试获取总览统计
func (suite *StatsServiceTestSuite) TestGetOverviewStats() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	
	// 创建2个分组
	group1 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP001")
	group2 := CreateTestGroup(suite.T(), TestDB, user.ID, "GROUP002")
	
	// 创建3个账号（2个在group1，1个在group2）
	account1 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "line1", "line")
	account2 := CreateTestLineAccount(suite.T(), TestDB, group1.ID, "line2", "line")
	account3 := CreateTestLineAccount(suite.T(), TestDB, group2.ID, "line3", "line_business")
	
	// 设置账号1为在线
	TestDB.Model(account1).Update("online_status", "online")
	
	// 创建一些进线记录
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, "line_id_1", false, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account1.ID, group1.ID, "line_id_2", false, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account2.ID, group1.ID, "line_id_3", true, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account3.ID, group2.ID, "line_id_4", false, "line_business")
	
	// 更新分组统计
	TestDB.Model(&models.GroupStats{}).Where("group_id = ?", group1.ID).Updates(map[string]interface{}{
		"total_incoming":     3,
		"today_incoming":     3,
		"duplicate_incoming": 1,
		"today_duplicate":    1,
	})
	TestDB.Model(&models.GroupStats{}).Where("group_id = ?", group2.ID).Updates(map[string]interface{}{
		"total_incoming":     1,
		"today_incoming":     1,
		"duplicate_incoming": 0,
		"today_duplicate":    0,
	})
	
	// 获取总览统计
	c := suite.createTestContext()
	stats, err := suite.statsService.GetOverviewStats(c)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), int64(2), stats["total_groups"], "应该有2个分组")
	assert.Equal(suite.T(), int64(3), stats["total_accounts"], "应该有3个账号")
	assert.Equal(suite.T(), int64(1), stats["online_accounts"], "应该有1个在线账号")
	assert.Equal(suite.T(), int64(4), stats["total_incoming"], "总进线数应该为4")
	assert.Equal(suite.T(), int64(4), stats["today_incoming"], "今日进线数应该为4")
	assert.Equal(suite.T(), int64(1), stats["duplicate_incoming"], "重复进线数应该为1")
	assert.Equal(suite.T(), int64(1), stats["today_duplicate"], "今日重复数应该为1")
}

// TestGetGroupIncomingTrend 测试获取分组进线趋势
func (suite *StatsServiceTestSuite) TestGetGroupIncomingTrend() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建不同日期的进线记录（需要截断到天，确保日期匹配）
	now := time.Now().Truncate(24 * time.Hour)
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)
	
	// 今天：2条进线，1条重复（在创建时直接设置时间，避免复合主键冲突）
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_today_1", false, "line", now)
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_today_2", true, "line", now)
	
	// 昨天：3条进线，0条重复
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_yesterday_1", false, "line", yesterday)
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_yesterday_2", false, "line", yesterday)
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_yesterday_3", false, "line", yesterday)
	
	// 前天：1条进线，1条重复
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_2days_1", true, "line", twoDaysAgo)
	
	// 获取7天趋势
	trend, err := suite.statsService.GetGroupIncomingTrend(group.ID, 7)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), trend)
	assert.Len(suite.T(), trend, 7, "应该返回7天的数据")
	
	// 找到今天的数据（通过日期匹配）
	todayStr := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")
	var todayData map[string]interface{}
	for _, dayData := range trend {
		if dayData["date"] == todayStr {
			todayData = dayData
			break
		}
	}
	assert.NotNil(suite.T(), todayData, "应该找到今天的数据")
	assert.Equal(suite.T(), int64(2), todayData["incoming_count"], "今天应该有2条进线")
	assert.Equal(suite.T(), int64(1), todayData["duplicate_count"], "今天应该有1条重复")
	
	// 找到昨天的数据（通过日期匹配）
	yesterdayStr := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour).Format("2006-01-02")
	var yesterdayData map[string]interface{}
	for _, dayData := range trend {
		if dayData["date"] == yesterdayStr {
			yesterdayData = dayData
			break
		}
	}
	assert.NotNil(suite.T(), yesterdayData, "应该找到昨天的数据")
	assert.Equal(suite.T(), int64(3), yesterdayData["incoming_count"], "昨天应该有3条进线")
	assert.Equal(suite.T(), int64(0), yesterdayData["duplicate_count"], "昨天应该有0条重复")
	
	// 找到前天的数据（通过日期匹配）
	twoDaysAgoStr := time.Now().AddDate(0, 0, -2).Truncate(24 * time.Hour).Format("2006-01-02")
	var twoDaysAgoData map[string]interface{}
	for _, dayData := range trend {
		if dayData["date"] == twoDaysAgoStr {
			twoDaysAgoData = dayData
			break
		}
	}
	assert.NotNil(suite.T(), twoDaysAgoData, "应该找到前天的数据")
	assert.Equal(suite.T(), int64(1), twoDaysAgoData["incoming_count"], "前天应该有1条进线")
	assert.Equal(suite.T(), int64(1), twoDaysAgoData["duplicate_count"], "前天应该有1条重复")
}

// TestGetAccountIncomingTrend 测试获取账号进线趋势
func (suite *StatsServiceTestSuite) TestGetAccountIncomingTrend() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建不同日期的进线记录（需要截断到天，确保日期匹配）
	now := time.Now().Truncate(24 * time.Hour)
	yesterday := now.AddDate(0, 0, -1)
	
	// 今天：1条进线，0条重复（在创建时直接设置时间，避免复合主键冲突）
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_account_today_1", false, "line", now)
	
	// 昨天：2条进线，1条重复
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_account_yesterday_1", false, "line", yesterday)
	CreateTestIncomingLogWithTime(suite.T(), TestDB, account.ID, group.ID, "line_id_account_yesterday_2", true, "line", yesterday)
	
	// 获取7天趋势
	trend, err := suite.statsService.GetAccountIncomingTrend(account.ID, 7)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), trend)
	assert.Len(suite.T(), trend, 7, "应该返回7天的数据")
	
	// 找到今天的数据（通过日期匹配）
	todayStr := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")
	var todayData map[string]interface{}
	for _, dayData := range trend {
		if dayData["date"] == todayStr {
			todayData = dayData
			break
		}
	}
	assert.NotNil(suite.T(), todayData, "应该找到今天的数据")
	assert.Equal(suite.T(), int64(1), todayData["incoming_count"], "今天应该有1条进线")
	assert.Equal(suite.T(), int64(0), todayData["duplicate_count"], "今天应该有0条重复")
	
	// 找到昨天的数据（通过日期匹配）
	yesterdayStr := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour).Format("2006-01-02")
	var yesterdayData map[string]interface{}
	for _, dayData := range trend {
		if dayData["date"] == yesterdayStr {
			yesterdayData = dayData
			break
		}
	}
	assert.NotNil(suite.T(), yesterdayData, "应该找到昨天的数据")
	assert.Equal(suite.T(), int64(2), yesterdayData["incoming_count"], "昨天应该有2条进线")
	assert.Equal(suite.T(), int64(1), yesterdayData["duplicate_count"], "昨天应该有1条重复")
}

// TestGetOverviewStats_EmptyData 测试获取总览统计 - 空数据
func (suite *StatsServiceTestSuite) TestGetOverviewStats_EmptyData() {
	// 不创建任何数据
	
	// 获取总览统计
	c := suite.createTestContext()
	stats, err := suite.statsService.GetOverviewStats(c)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), int64(0), stats["total_groups"], "应该有0个分组")
	assert.Equal(suite.T(), int64(0), stats["total_accounts"], "应该有0个账号")
	assert.Equal(suite.T(), int64(0), stats["online_accounts"], "应该有0个在线账号")
	assert.Equal(suite.T(), int64(0), stats["total_incoming"], "总进线数应该为0")
	assert.Equal(suite.T(), int64(0), stats["today_incoming"], "今日进线数应该为0")
	assert.Equal(suite.T(), int64(0), stats["duplicate_incoming"], "重复进线数应该为0")
	assert.Equal(suite.T(), int64(0), stats["today_duplicate"], "今日重复数应该为0")
}

// TestGetGroupIncomingTrend_NoData 测试获取分组进线趋势 - 无数据
func (suite *StatsServiceTestSuite) TestGetGroupIncomingTrend_NoData() {
	// 创建测试数据（但不创建进线记录）
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	
	// 获取7天趋势
	trend, err := suite.statsService.GetGroupIncomingTrend(group.ID, 7)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), trend)
	assert.Len(suite.T(), trend, 7, "应该返回7天的数据")
	
	// 验证所有天数的数据都是0
	for _, dayData := range trend {
		assert.Equal(suite.T(), int64(0), dayData["incoming_count"], "进线数应该为0")
		assert.Equal(suite.T(), int64(0), dayData["duplicate_count"], "重复数应该为0")
	}
}

// TestGetAccountIncomingTrend_DifferentDays 测试获取账号进线趋势 - 不同天数
func (suite *StatsServiceTestSuite) TestGetAccountIncomingTrend_DifferentDays() {
	// 创建测试数据
	user := CreateTestUser(suite.T(), TestDB, "admin")
	group := CreateTestGroup(suite.T(), TestDB, user.ID, "")
	account := CreateTestLineAccount(suite.T(), TestDB, group.ID, "", "line")
	
	// 创建今天的进线记录
	now := time.Now()
	log1 := CreateTestIncomingLog(suite.T(), TestDB, account.ID, group.ID, "line_id_test_1", false, "line")
	log1.IncomingTime = now
	TestDB.Save(log1)
	
	// 测试不同天数的趋势
	for _, days := range []int{7, 15, 30} {
		trend, err := suite.statsService.GetAccountIncomingTrend(account.ID, days)
		
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), trend)
		assert.Len(suite.T(), trend, days, "应该返回%d天的数据", days)
	}
}

// TestGetOverviewStats_UserFilter 测试获取总览统计 - 用户过滤
func (suite *StatsServiceTestSuite) TestGetOverviewStats_UserFilter() {
	// 创建两个用户
	user1 := CreateTestUser(suite.T(), TestDB, "user")
	user2 := CreateTestUser(suite.T(), TestDB, "user")

	// 为user1创建2个分组，为user2创建1个分组
	group1_1 := CreateTestGroup(suite.T(), TestDB, user1.ID, "GROUP1_1")
	group1_2 := CreateTestGroup(suite.T(), TestDB, user1.ID, "GROUP1_2")
	group2_1 := CreateTestGroup(suite.T(), TestDB, user2.ID, "GROUP2_1")

	// 为每个分组创建账号
	account1_1 := CreateTestLineAccount(suite.T(), TestDB, group1_1.ID, "line1_1", "line")
	account1_2 := CreateTestLineAccount(suite.T(), TestDB, group1_2.ID, "line1_2", "line")
	account2_1 := CreateTestLineAccount(suite.T(), TestDB, group2_1.ID, "line2_1", "line")

	// 设置账号状态
	TestDB.Model(account1_1).Update("online_status", "online")
	TestDB.Model(account2_1).Update("online_status", "online")

	// 创建进线记录
	CreateTestIncomingLog(suite.T(), TestDB, account1_1.ID, group1_1.ID, "line_id_1", false, "line")
	CreateTestIncomingLog(suite.T(), TestDB, account1_2.ID, group1_2.ID, "line_id_2", true, "line")

	// 创建联系人记录
	CreateTestContactPool(suite.T(), TestDB, group1_1.ID, "contact1", "line")
	CreateTestContactPool(suite.T(), TestDB, group1_2.ID, "contact2", "line")

	// 测试管理员权限（可以看到所有数据）
	adminStats, err := suite.statsService.GetOverviewStats(suite.createTestContext())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), adminStats["total_groups"], "管理员应该看到3个分组")
	assert.Equal(suite.T(), int64(3), adminStats["total_accounts"], "管理员应该看到3个账号")
	assert.Equal(suite.T(), int64(2), adminStats["online_accounts"], "管理员应该看到2个在线账号")
	assert.Equal(suite.T(), int64(2), adminStats["total_incoming"], "管理员应该看到2条进线")
	assert.Equal(suite.T(), int64(2), adminStats["total_contacts"], "管理员应该看到2条联系人")

	// 测试普通用户user1的权限（只能看到自己的数据）
	user1Stats, err := suite.statsService.GetOverviewStats(suite.createUserContext(user1.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), user1Stats["total_groups"], "user1应该看到2个分组")
	assert.Equal(suite.T(), int64(2), user1Stats["total_accounts"], "user1应该看到2个账号")
	assert.Equal(suite.T(), int64(1), user1Stats["online_accounts"], "user1应该看到1个在线账号")
	assert.Equal(suite.T(), int64(2), user1Stats["total_incoming"], "user1应该看到2条进线")
	assert.Equal(suite.T(), int64(2), user1Stats["total_contacts"], "user1应该看到2条联系人")

	// 测试普通用户user2的权限（只能看到自己的数据）
	user2Stats, err := suite.statsService.GetOverviewStats(suite.createUserContext(user2.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), user2Stats["total_groups"], "user2应该看到1个分组")
	assert.Equal(suite.T(), int64(1), user2Stats["total_accounts"], "user2应该看到1个账号")
	assert.Equal(suite.T(), int64(1), user2Stats["online_accounts"], "user2应该看到1个在线账号")
	assert.Equal(suite.T(), int64(0), user2Stats["total_incoming"], "user2应该看到0条进线")
	assert.Equal(suite.T(), int64(0), user2Stats["total_contacts"], "user2应该看到0条联系人")

	// 测试子账号权限（不能查看分组数据，其他数据按分组过滤）
	subAccountStats, err := suite.statsService.GetOverviewStats(suite.createSubAccountContext(group1_1.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), subAccountStats["total_groups"], "子账号应该看到0个分组")
	assert.Equal(suite.T(), int64(1), subAccountStats["total_accounts"], "子账号应该看到1个账号")
	assert.Equal(suite.T(), int64(1), subAccountStats["online_accounts"], "子账号应该看到1个在线账号")
	assert.Equal(suite.T(), int64(1), subAccountStats["total_incoming"], "子账号应该看到1条进线")
	assert.Equal(suite.T(), int64(1), subAccountStats["total_contacts"], "子账号应该看到1条联系人")
}

// TestStatsServiceTestSuite 运行测试套件
func TestStatsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StatsServiceTestSuite))
}

