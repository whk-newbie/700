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

// CustomerServiceTestSuite 客户服务测试套件
type CustomerServiceTestSuite struct {
	suite.Suite
	db            *gorm.DB
	service       *services.CustomerService
	adminUser     *models.User
	testGroup     *models.Group
	testAccount   *models.LineAccount
	testCustomer  *models.Customer
}

// SetupSuite 测试套件初始化
func (suite *CustomerServiceTestSuite) SetupSuite() {
	suite.db = SetupTestDB(suite.T())
	suite.service = services.NewCustomerService()
	
	// 创建测试用户、分组和账号
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
	suite.testAccount = CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_001", "line")
	
	// 创建测试客户
	suite.testCustomer = &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		LineAccountID:  &suite.testAccount.ID,
		PlatformType:   "line",
		CustomerID:     "customer_001",
		DisplayName:    "测试客户",
		PhoneNumber:    "13800138000",
		CustomerType:   "realtime",
		Gender:         "unknown", // 设置为有效值，避免违反检查约束
	}
	err := suite.db.Create(suite.testCustomer).Error
	if err != nil {
		suite.T().Fatalf("Failed to create test customer: %v", err)
	}
}

// TearDownTest 每个测试后清理
func (suite *CustomerServiceTestSuite) TearDownTest() {
	CleanupTestData(suite.T(), suite.db)
	// 重新创建基础数据
	suite.adminUser = CreateTestUser(suite.T(), suite.db, "admin")
	suite.testGroup = CreateTestGroup(suite.T(), suite.db, suite.adminUser.ID, "TEST0001")
	suite.testAccount = CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_001", "line")
	suite.testCustomer = &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		LineAccountID:  &suite.testAccount.ID,
		PlatformType:   "line",
		CustomerID:     "customer_001",
		DisplayName:    "测试客户",
		PhoneNumber:    "13800138000",
		CustomerType:   "realtime",
		Gender:         "unknown", // 设置为有效值，避免违反检查约束
	}
	err := suite.db.Create(suite.testCustomer).Error
	if err != nil {
		suite.T().Fatalf("Failed to create test customer: %v", err)
	}
}

// TestCreateCustomer_Success 测试成功创建客户
func (suite *CustomerServiceTestSuite) TestCreateCustomer_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateCustomerRequest{
		GroupID:      suite.testGroup.ID,
		LineAccountID: &suite.testAccount.ID,
		PlatformType:  "line",
		CustomerID:    "customer_002",
		DisplayName:   "测试客户2",
		PhoneNumber:   "13900139000",
		CustomerType:  "realtime",
		Gender:        "unknown", // 设置为有效值
	}
	
	customer, err := suite.service.CreateCustomer(c, req)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.Equal(suite.testGroup.ID, customer.GroupID)
	suite.Equal("customer_002", customer.CustomerID)
	suite.Equal("line", customer.PlatformType)
	suite.Equal("realtime", customer.CustomerType)
}

// TestCreateCustomer_GroupNotExists 测试分组不存在
func (suite *CustomerServiceTestSuite) TestCreateCustomer_GroupNotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateCustomerRequest{
		GroupID:      99999, // 不存在的分组ID
		PlatformType: "line",
		CustomerID:   "customer_003",
	}
	
	customer, err := suite.service.CreateCustomer(c, req)
	suite.Error(err)
	suite.Nil(customer)
	suite.Contains(err.Error(), "分组不存在")
}

// TestCreateCustomer_LineAccountNotExists 测试Line账号不存在
func (suite *CustomerServiceTestSuite) TestCreateCustomer_LineAccountNotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	nonExistentID := uint(99999)
	req := &schemas.CreateCustomerRequest{
		GroupID:      suite.testGroup.ID,
		LineAccountID: &nonExistentID, // 不存在的账号ID
		PlatformType:  "line",
		CustomerID:    "customer_004",
	}
	
	customer, err := suite.service.CreateCustomer(c, req)
	suite.Error(err)
	suite.Nil(customer)
	suite.Contains(err.Error(), "Line账号不存在")
}

// TestCreateCustomer_DuplicateCustomerID 测试重复的客户ID
func (suite *CustomerServiceTestSuite) TestCreateCustomer_DuplicateCustomerID() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateCustomerRequest{
		GroupID:      suite.testGroup.ID,
		PlatformType:  "line",
		CustomerID:    suite.testCustomer.CustomerID, // 使用已存在的客户ID
	}
	
	customer, err := suite.service.CreateCustomer(c, req)
	suite.Error(err)
	suite.Nil(customer)
	suite.Contains(err.Error(), "该客户在此分组中已存在")
}

// TestCreateCustomer_WithBirthday 测试创建带生日的客户
func (suite *CustomerServiceTestSuite) TestCreateCustomer_WithBirthday() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.CreateCustomerRequest{
		GroupID:      suite.testGroup.ID,
		PlatformType:  "line",
		CustomerID:    "customer_005",
		DisplayName:   "测试客户5",
		Birthday:      "1990-01-01",
		Gender:        "unknown", // 设置为有效值
	}
	
	customer, err := suite.service.CreateCustomer(c, req)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.NotNil(customer.Birthday)
	suite.Equal("1990-01-01", customer.Birthday.Format("2006-01-02"))
}

// TestGetCustomerList_Success 测试获取客户列表
func (suite *CustomerServiceTestSuite) TestGetCustomerList_Success() {
	// 创建多个测试客户
	customer1 := &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		PlatformType:   "line",
		CustomerID:     "customer_006",
		DisplayName:    "测试客户6",
		CustomerType:   "realtime",
		Gender:         "unknown",
	}
	suite.db.Create(customer1)
	
	customer2 := &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		PlatformType:   "line_business",
		CustomerID:     "customer_007",
		DisplayName:    "测试客户7",
		CustomerType:   "supplement",
		Gender:         "unknown",
	}
	suite.db.Create(customer2)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.CustomerQueryParams{
		Page:     1,
		PageSize: 10,
	}
	
	list, total, err := suite.service.GetCustomerList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(3)) // 至少包含suite.testCustomer和刚创建的两个
	suite.GreaterOrEqual(len(list), 3)
}

// TestGetCustomerList_WithFilter 测试带筛选的列表查询
func (suite *CustomerServiceTestSuite) TestGetCustomerList_WithFilter() {
	// 创建不同平台和类型的客户
	lineCustomer := &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		PlatformType:   "line",
		CustomerID:     "customer_008",
		DisplayName:    "Line客户",
		CustomerType:   "realtime",
		Gender:         "unknown",
	}
	suite.db.Create(lineCustomer)
	
	businessCustomer := &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		PlatformType:   "line_business",
		CustomerID:     "customer_009",
		DisplayName:    "Business客户",
		CustomerType:   "supplement",
		Gender:         "unknown",
	}
	suite.db.Create(businessCustomer)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	// 筛选line平台
	params := &schemas.CustomerQueryParams{
		Page:         1,
		PageSize:     10,
		PlatformType: "line",
	}
	
	list, total, err := suite.service.GetCustomerList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	for _, cust := range list {
		suite.Equal("line", cust.PlatformType)
	}
	
	// 筛选客户类型
	params.PlatformType = ""
	params.CustomerType = "realtime"
	list2, total2, err := suite.service.GetCustomerList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total2, int64(1))
	for _, cust := range list2 {
		suite.Equal("realtime", cust.CustomerType)
	}
}

// TestGetCustomerList_WithSearch 测试搜索功能
func (suite *CustomerServiceTestSuite) TestGetCustomerList_WithSearch() {
	// 创建带特定显示名称的客户
	customer := &models.Customer{
		GroupID:        suite.testGroup.ID,
		ActivationCode: suite.testGroup.ActivationCode,
		PlatformType:   "line",
		CustomerID:     "customer_010",
		DisplayName:    "特殊名称客户",
		CustomerType:   "realtime",
		Gender:         "unknown",
	}
	suite.db.Create(customer)
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	params := &schemas.CustomerQueryParams{
		Page:     1,
		PageSize: 10,
		Search:   "特殊名称",
	}
	
	list, total, err := suite.service.GetCustomerList(c, params)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	
	found := false
	for _, cust := range list {
		if cust.ID == customer.ID {
			found = true
			break
		}
	}
	suite.True(found, "应该找到包含'特殊名称'的客户")
}

// TestGetCustomerDetail_Success 测试获取客户详情
func (suite *CustomerServiceTestSuite) TestGetCustomerDetail_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	detail, err := suite.service.GetCustomerDetail(c, suite.testCustomer.ID)
	suite.NoError(err)
	suite.NotNil(detail)
	suite.Equal(suite.testCustomer.ID, detail.ID)
	suite.Equal(suite.testCustomer.CustomerID, detail.CustomerID)
}

// TestGetCustomerDetail_NotExists 测试客户不存在
func (suite *CustomerServiceTestSuite) TestGetCustomerDetail_NotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	detail, err := suite.service.GetCustomerDetail(c, 99999)
	suite.Error(err)
	suite.Nil(detail)
	suite.Contains(err.Error(), "客户不存在")
}

// TestUpdateCustomer_Success 测试更新客户
func (suite *CustomerServiceTestSuite) TestUpdateCustomer_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.UpdateCustomerRequest{
		DisplayName:  "更新后的名称",
		PhoneNumber:  "13900139000",
		CustomerType: "supplement",
		Gender:       "male",
		Country:      "CN",
		Birthday:     "1990-01-01",
		Address:      "测试地址",
		Remark:       "更新后的备注",
	}
	
	customer, err := suite.service.UpdateCustomer(c, suite.testCustomer.ID, req)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.Equal("更新后的名称", customer.DisplayName)
	suite.Equal("13900139000", customer.PhoneNumber)
	suite.Equal("supplement", customer.CustomerType)
	suite.Equal("male", customer.Gender)
	suite.Equal("CN", customer.Country)
	suite.NotNil(customer.Birthday)
	suite.Equal("1990-01-01", customer.Birthday.Format("2006-01-02"))
	suite.Equal("测试地址", customer.Address)
	suite.Equal("更新后的备注", customer.Remark)
}

// TestUpdateCustomer_ChangeLineAccount 测试更换Line账号
func (suite *CustomerServiceTestSuite) TestUpdateCustomer_ChangeLineAccount() {
	// 创建新的Line账号
	newAccount := CreateTestLineAccount(suite.T(), suite.db, suite.testGroup.ID, "test_line_002", "line")
	
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	req := &schemas.UpdateCustomerRequest{
		LineAccountID: &newAccount.ID,
	}
	
	customer, err := suite.service.UpdateCustomer(c, suite.testCustomer.ID, req)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.Equal(newAccount.ID, *customer.LineAccountID)
}

// TestUpdateCustomer_LineAccountNotExists 测试Line账号不存在
func (suite *CustomerServiceTestSuite) TestUpdateCustomer_LineAccountNotExists() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	nonExistentID := uint(99999)
	req := &schemas.UpdateCustomerRequest{
		LineAccountID: &nonExistentID,
	}
	
	customer, err := suite.service.UpdateCustomer(c, suite.testCustomer.ID, req)
	suite.Error(err)
	suite.Nil(customer)
	suite.Contains(err.Error(), "Line账号不存在")
}

// TestDeleteCustomer_Success 测试删除客户（软删除）
func (suite *CustomerServiceTestSuite) TestDeleteCustomer_Success() {
	c, _ := gin.CreateTestContext(nil)
	c.Set("user_id", suite.adminUser.ID)
	c.Set("role", "admin")
	
	err := suite.service.DeleteCustomer(c, suite.testCustomer.ID)
	suite.NoError(err)
	
	// 验证软删除（应该查询不到）
	var customer models.Customer
	err = suite.db.Where("id = ? AND deleted_at IS NULL", suite.testCustomer.ID).First(&customer).Error
	suite.Error(err)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound))
	
	// 验证硬删除查询（应该能查到）
	err = suite.db.Unscoped().Where("id = ?", suite.testCustomer.ID).First(&customer).Error
	suite.NoError(err)
	suite.NotNil(customer.DeletedAt)
}

// TestSyncCustomer_CreateNew 测试同步客户（创建新客户）
func (suite *CustomerServiceTestSuite) TestSyncCustomer_CreateNew() {
	data := &schemas.CustomerSyncData{
		CustomerID:   "customer_sync_001",
		PlatformType: "line",
		DisplayName:  "同步客户1",
		PhoneNumber:  "13800138001",
		Gender:       "female",
		Country:      "JP",
		Address:      "同步地址",
		Birthday:     "1995-05-15",
		Remark:       "同步备注",
	}
	
	customer, err := suite.service.SyncCustomer(suite.testGroup.ID, suite.testGroup.ActivationCode, data)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.Equal("customer_sync_001", customer.CustomerID)
	suite.Equal("同步客户1", customer.DisplayName)
	suite.Equal("13800138001", customer.PhoneNumber)
	suite.Equal("female", customer.Gender)
	suite.Equal("JP", customer.Country)
	suite.Equal("同步地址", customer.Address)
	suite.Equal("同步备注", customer.Remark)
	suite.NotNil(customer.Birthday)
	suite.Equal("1995-05-15", customer.Birthday.Format("2006-01-02"))
}

// TestSyncCustomer_UpdateExisting 测试同步客户（更新现有客户）
func (suite *CustomerServiceTestSuite) TestSyncCustomer_UpdateExisting() {
	data := &schemas.CustomerSyncData{
		CustomerID:   suite.testCustomer.CustomerID, // 使用已存在的客户ID
		PlatformType: suite.testCustomer.PlatformType,
		DisplayName:  "更新后的同步名称",
		PhoneNumber:  "13900139001",
		Gender:       "male",
		Country:      "US",
		Address:      "更新后的地址",
		Remark:       "更新后的备注",
	}
	
	customer, err := suite.service.SyncCustomer(suite.testGroup.ID, suite.testGroup.ActivationCode, data)
	suite.NoError(err)
	suite.NotNil(customer)
	suite.Equal(suite.testCustomer.ID, customer.ID) // 应该是同一个客户
	suite.Equal("更新后的同步名称", customer.DisplayName)
	suite.Equal("13900139001", customer.PhoneNumber)
	suite.Equal("male", customer.Gender)
	suite.Equal("US", customer.Country)
	suite.Equal("更新后的地址", customer.Address)
	suite.Equal("更新后的备注", customer.Remark)
}

// TestCustomerServiceTestSuite 运行测试套件
func TestCustomerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CustomerServiceTestSuite))
}

