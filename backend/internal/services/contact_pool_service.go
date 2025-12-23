package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"line-management/internal/models"
	"line-management/internal/schemas"
	"line-management/internal/utils"
	"line-management/pkg/database"
	"line-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ContactPoolService 底库服务
type ContactPoolService struct {
	db *gorm.DB
}

// NewContactPoolService 创建底库服务实例
func NewContactPoolService() *ContactPoolService {
	return &ContactPoolService{
		db: database.GetDB(),
	}
}

// GetSummary 获取底库统计汇总
func (s *ContactPoolService) GetSummary(c *gin.Context) (*schemas.ContactPoolSummaryResponse, error) {
	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.ContactPool{}), "contact_pool")

	// 统计导入的联系人数量（source_type = 'import'）
	var importCount int64
	importQuery := query.Where("contact_pool.source_type = ? AND contact_pool.deleted_at IS NULL", "import")
	if err := importQuery.Count(&importCount).Error; err != nil {
		logger.Errorf("统计导入联系人数量失败: %v", err)
		return nil, err
	}

	// 统计平台工单联系人数量（source_type = 'platform'）
	var platformCount int64
	platformQuery := query.Where("contact_pool.source_type = ? AND contact_pool.deleted_at IS NULL", "platform")
	if err := platformQuery.Count(&platformCount).Error; err != nil {
		logger.Errorf("统计平台工单联系人数量失败: %v", err)
		return nil, err
	}

	return &schemas.ContactPoolSummaryResponse{
		ImportCount:   importCount,
		PlatformCount: platformCount,
		TotalCount:    importCount + platformCount,
	}, nil
}

// GetList 获取底库列表（按激活码+平台）
func (s *ContactPoolService) GetList(c *gin.Context, params *schemas.ContactPoolListQueryParams) ([]schemas.ContactPoolListResponse, int64, error) {
	// 设置默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}

	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.ContactPool{}), "contact_pool")
	
	// 只查询平台工单来源的数据（source_type = 'platform'）
	query = query.Where("contact_pool.source_type = ? AND contact_pool.deleted_at IS NULL", "platform")

	// 筛选条件
	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}
	if params.Search != "" {
		query = query.Where("activation_code LIKE ?", "%"+params.Search+"%")
	}

	// 按激活码和平台分组统计
	type GroupResult struct {
		ActivationCode string
		PlatformType   string
		ContactCount   int64
	}

	var results []GroupResult
	if err := query.
		Select("contact_pool.activation_code, contact_pool.platform_type, COUNT(*) as contact_count").
		Group("contact_pool.activation_code, contact_pool.platform_type").
		Scan(&results).Error; err != nil {
		logger.Errorf("查询底库列表失败: %v", err)
		return nil, 0, err
	}

	// 获取分组备注信息
	type ResultWithRemark struct {
		ActivationCode string
		Remark         string
		PlatformType   string
		ContactCount   int64
	}
	var resultsWithRemark []ResultWithRemark
	for _, r := range results {
		var group models.Group
		remark := ""
		if err := s.db.Where("activation_code = ? AND deleted_at IS NULL", r.ActivationCode).
			First(&group).Error; err == nil {
			remark = group.Remark
		}
		resultsWithRemark = append(resultsWithRemark, ResultWithRemark{
			ActivationCode: r.ActivationCode,
			Remark:         remark,
			PlatformType:   r.PlatformType,
			ContactCount:   r.ContactCount,
		})
	}

	// 转换为响应格式
	var list []schemas.ContactPoolListResponse
	for _, r := range resultsWithRemark {
		list = append(list, schemas.ContactPoolListResponse{
			ActivationCode: r.ActivationCode,
			Remark:         r.Remark,
			PlatformType:   r.PlatformType,
			ContactCount:   r.ContactCount,
		})
	}

	// 计算总数
	total := int64(len(resultsWithRemark))

	// 分页（简单分页，因为已经分组了）
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start > len(list) {
		list = []schemas.ContactPoolListResponse{}
	} else if end > len(list) {
		list = list[start:]
	} else {
		list = list[start:end]
	}

	return list, total, nil
}

// GetDetailList 获取底库详细列表
func (s *ContactPoolService) GetDetailList(c *gin.Context, params *schemas.ContactPoolDetailQueryParams) ([]schemas.ContactPoolDetailResponse, int64, error) {
	// 设置默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}

	// 应用数据过滤
	query := utils.ApplyDataFilter(c, s.db.Model(&models.ContactPool{}), "contact_pool")
	query = query.Where("contact_pool.deleted_at IS NULL")

	// 筛选条件
	if params.ActivationCode != "" {
		query = query.Where("activation_code = ?", params.ActivationCode)
	}
	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("(line_id LIKE ? OR display_name LIKE ? OR phone_number LIKE ?)", 
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("统计底库详细列表总数失败: %v", err)
		return nil, 0, err
	}

	// 分页查询
	var contacts []models.ContactPool
	offset := (params.Page - 1) * params.PageSize
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&contacts).Error; err != nil {
		logger.Errorf("查询底库详细列表失败: %v", err)
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.ContactPoolDetailResponse
	for _, contact := range contacts {
		source := "系统上报"
		if contact.SourceType == "import" {
			source = "手动导入"
		}

		list = append(list, schemas.ContactPoolDetailResponse{
			ID:          contact.ID,
			LineID:      contact.LineID,
			DisplayName: contact.DisplayName,
			PhoneNumber: contact.PhoneNumber,
			Source:      source,
			CreatedAt:   contact.CreatedAt,
		})
	}

	return list, total, nil
}

// ImportContacts 导入联系人（从文件）
func (s *ContactPoolService) ImportContacts(c *gin.Context, file *multipart.FileHeader, req *schemas.ImportContactRequest) (*schemas.ImportContactResponse, error) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("无法获取用户信息")
	}
	userIDUint := userID.(uint)
	var importedBy *uint
	if userIDUint > 0 {
		importedBy = &userIDUint
	}

	// 验证分组是否存在
	var group models.Group
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.GroupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	// 创建上传目录
	uploadDir := "./static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Errorf("创建上传目录失败: %v", err)
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 保存文件
	fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, fileName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Errorf("保存文件失败: %v", err)
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 创建导入批次记录
	batch := models.ImportBatch{
		BatchName:    file.Filename,
		PlatformType: req.PlatformType,
		DedupScope:   req.DedupScope,
		FileName:     file.Filename,
		FilePath:     filePath,
		FileSize:     file.Size,
		ImportedBy:   importedBy,
	}

	if err := s.db.Create(&batch).Error; err != nil {
		os.Remove(filePath) // 删除已保存的文件
		logger.Errorf("创建导入批次失败: %v", err)
		return nil, fmt.Errorf("创建导入批次失败: %v", err)
	}

	// 解析文件
	contacts, err := s.parseFile(filePath, file.Filename)
	if err != nil {
		// 更新批次状态
		s.db.Model(&batch).Updates(map[string]interface{}{
			"error_count": len(contacts),
			"total_count": len(contacts),
		})
		os.Remove(filePath)
		return nil, fmt.Errorf("解析文件失败: %v", err)
	}

	// 批量插入联系人
	result := s.batchInsertContacts(c, contacts, &batch, req.GroupID, group.ActivationCode, req.PlatformType, req.DedupScope)

	// 更新批次状态
	now := time.Now()
	s.db.Model(&batch).Updates(map[string]interface{}{
		"total_count":    result.TotalCount,
		"success_count":  result.SuccessCount,
		"duplicate_count": result.DuplicateCount,
		"error_count":    result.ErrorCount,
		"completed_at":   &now,
	})

	return &schemas.ImportContactResponse{
		BatchID:       batch.ID,
		TotalCount:    result.TotalCount,
		SuccessCount:  result.SuccessCount,
		DuplicateCount: result.DuplicateCount,
		ErrorCount:    result.ErrorCount,
	}, nil
}

// parseFile 解析文件（支持Excel和CSV）
func (s *ContactPoolService) parseFile(filePath, fileName string) ([]ContactRow, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	
	switch ext {
	case ".xlsx", ".xls":
		return s.parseExcel(filePath)
	case ".csv":
		return s.parseCSV(filePath)
	case ".txt":
		return s.parseTXT(filePath)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// ContactRow 联系人行数据
type ContactRow struct {
	LineID      string
	DisplayName string
	PhoneNumber string
	Remark      string
}

// parseExcel 解析Excel文件
func (s *ContactPoolService) parseExcel(filePath string) ([]ContactRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Excel文件没有工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取Excel行失败: %v", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel文件至少需要包含标题行和一行数据")
	}

	var contacts []ContactRow
	// 跳过第一行（标题行），从第二行开始读取
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		// 至少需要Line ID（第一列）
		if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		contact := ContactRow{
			LineID: strings.TrimSpace(row[0]),
		}

		// 显示名称（第二列，可选）
		if len(row) > 1 {
			contact.DisplayName = strings.TrimSpace(row[1])
		}

		// 手机号（第三列，可选）
		if len(row) > 2 {
			contact.PhoneNumber = strings.TrimSpace(row[2])
		}

		// 备注（第四列，可选）
		if len(row) > 3 {
			contact.Remark = strings.TrimSpace(row[3])
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// parseCSV 解析CSV文件
func (s *ContactPoolService) parseCSV(filePath string) ([]ContactRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV文件失败: %v", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("CSV文件至少需要包含标题行和一行数据")
	}

	var contacts []ContactRow
	// 跳过第一行（标题行），从第二行开始读取
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		// 至少需要Line ID（第一列）
		if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		contact := ContactRow{
			LineID: strings.TrimSpace(row[0]),
		}

		// 显示名称（第二列，可选）
		if len(row) > 1 {
			contact.DisplayName = strings.TrimSpace(row[1])
		}

		// 手机号（第三列，可选）
		if len(row) > 2 {
			contact.PhoneNumber = strings.TrimSpace(row[2])
		}

		// 备注（第四列，可选）
		if len(row) > 3 {
			contact.Remark = strings.TrimSpace(row[3])
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// parseTXT 解析TXT文件（每行一个Line ID，可选的其他字段用制表符或逗号分隔）
func (s *ContactPoolService) parseTXT(filePath string) ([]ContactRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开TXT文件失败: %v", err)
	}
	defer file.Close()

	var contacts []ContactRow
	reader := csv.NewReader(file)
	reader.Comma = '\t' // 尝试使用制表符分隔
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		// 如果制表符分隔失败，尝试逗号分隔
		file.Seek(0, 0)
		reader = csv.NewReader(file)
		reader.Comma = ','
		rows, err = reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("读取TXT文件失败: %v", err)
		}
	}

	for _, row := range rows {
		if len(row) == 0 {
			continue
		}

		// 至少需要Line ID（第一列）
		if strings.TrimSpace(row[0]) == "" {
			continue
		}

		contact := ContactRow{
			LineID: strings.TrimSpace(row[0]),
		}

		// 显示名称（第二列，可选）
		if len(row) > 1 {
			contact.DisplayName = strings.TrimSpace(row[1])
		}

		// 手机号（第三列，可选）
		if len(row) > 2 {
			contact.PhoneNumber = strings.TrimSpace(row[2])
		}

		// 备注（第四列，可选）
		if len(row) > 3 {
			contact.Remark = strings.TrimSpace(row[3])
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// BatchInsertResult 批量插入结果
type BatchInsertResult struct {
	TotalCount    int
	SuccessCount  int
	DuplicateCount int
	ErrorCount    int
}

// batchInsertContacts 批量插入联系人
func (s *ContactPoolService) batchInsertContacts(
	c *gin.Context,
	contacts []ContactRow,
	batch *models.ImportBatch,
	groupID uint,
	activationCode string,
	platformType string,
	dedupScope string,
) *BatchInsertResult {
	result := &BatchInsertResult{
		TotalCount: len(contacts),
	}

	dedupService := NewDedupService()
	now := time.Now()

	// 批量插入（分批处理，每批100条）
	batchSize := 100
	for i := 0; i < len(contacts); i += batchSize {
		end := i + batchSize
		if end > len(contacts) {
			end = len(contacts)
		}

		batchContacts := contacts[i:end]
		var toInsert []models.ContactPool

		for _, row := range batchContacts {
			// 检查是否重复
			isDuplicate := false
			if dedupScope == "global" {
				// 全局去重：检查底库和进线记录
				exists, _ := dedupService.CheckContactPoolDuplicate(row.LineID, platformType)
				if !exists {
					// 也检查进线记录
					dup, _ := dedupService.CheckDuplicateGlobal(row.LineID)
					exists = dup
				}
				isDuplicate = exists
			} else {
				// 当前分组去重：检查当前分组的进线记录
				dup, _ := dedupService.CheckDuplicateCurrent(groupID, row.LineID)
				isDuplicate = dup
			}

			if isDuplicate {
				result.DuplicateCount++
				continue
			}

			// 检查底库中是否已存在（避免重复插入）
			exists, _ := dedupService.CheckContactPoolDuplicate(row.LineID, platformType)
			if exists {
				result.DuplicateCount++
				continue
			}

			contact := models.ContactPool{
				SourceType:    "import",
				ImportBatchID: &batch.ID,
				GroupID:       groupID,
				ActivationCode: activationCode,
				PlatformType:  platformType,
				LineID:        row.LineID,
				DisplayName:   row.DisplayName,
				PhoneNumber:   row.PhoneNumber,
				DedupScope:    dedupScope,
				FirstSeenAt:   &now,
				Remark:        row.Remark,
			}

			toInsert = append(toInsert, contact)
		}

		// 批量插入
		if len(toInsert) > 0 {
			if err := s.db.CreateInBatches(toInsert, batchSize).Error; err != nil {
				logger.Errorf("批量插入联系人失败: %v", err)
				result.ErrorCount += len(toInsert)
			} else {
				result.SuccessCount += len(toInsert)
			}
		}
	}

	return result
}

// GetImportBatchList 获取导入批次列表
func (s *ContactPoolService) GetImportBatchList(c *gin.Context, params *schemas.ImportBatchListQueryParams) ([]schemas.ImportBatchListResponse, int64, error) {
	// 设置默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}

	// 应用数据过滤（根据用户角色过滤）
	query := utils.ApplyDataFilter(c, s.db.Model(&models.ImportBatch{}), "import_batches")

	// 筛选条件
	if params.PlatformType != "" {
		query = query.Where("platform_type = ?", params.PlatformType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("统计导入批次总数失败: %v", err)
		return nil, 0, err
	}

	// 分页查询
	var batches []models.ImportBatch
	offset := (params.Page - 1) * params.PageSize
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&batches).Error; err != nil {
		logger.Errorf("查询导入批次列表失败: %v", err)
		return nil, 0, err
	}

	// 转换为响应格式
	var list []schemas.ImportBatchListResponse
	for _, batch := range batches {
		list = append(list, schemas.ImportBatchListResponse{
			ID:            batch.ID,
			BatchName:     batch.BatchName,
			PlatformType:  batch.PlatformType,
			TotalCount:    batch.TotalCount,
			SuccessCount:  batch.SuccessCount,
			DuplicateCount: batch.DuplicateCount,
			ErrorCount:    batch.ErrorCount,
			DedupScope:    batch.DedupScope,
			FileName:      batch.FileName,
			CreatedAt:     batch.CreatedAt,
			CompletedAt:   batch.CompletedAt,
		})
	}

	return list, total, nil
}

// GenerateImportTemplate 生成导入模板文件
func (s *ContactPoolService) GenerateImportTemplate() (*excelize.File, error) {
	// 创建新的Excel文件
	f := excelize.NewFile()

	// 设置工作表名称
	sheetName := "联系人导入模板"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("创建工作表失败: %v", err)
	}

	// 删除默认的Sheet1
	f.DeleteSheet("Sheet1")

	// 设置活动工作表
	f.SetActiveSheet(index)

	// 设置表头
	headers := []string{"Line ID", "显示名称", "手机号", "备注"}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建表头样式失败: %v", err)
	}

	// 写入表头
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 25) // Line ID
	f.SetColWidth(sheetName, "B", "B", 20) // 显示名称
	f.SetColWidth(sheetName, "C", "C", 15) // 手机号
	f.SetColWidth(sheetName, "D", "D", 30) // 备注

	// 设置行高
	f.SetRowHeight(sheetName, 1, 25)

	// 添加示例数据行（可选）
	exampleStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建示例行样式失败: %v", err)
	}

	// 添加一行示例数据（灰色，作为提示）
	exampleData := []string{"U1234567890abcdefghijklmnopqrstuv", "示例联系人", "13800138000", "这是备注信息"}
	exampleFontStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#999999",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
		},
	})
	if err == nil {
		for i, data := range exampleData {
			cell := fmt.Sprintf("%c2", 'A'+i)
			f.SetCellValue(sheetName, cell, data)
			f.SetCellStyle(sheetName, cell, cell, exampleFontStyle)
		}
	} else {
		// 如果创建字体样式失败，使用普通样式
		for i, data := range exampleData {
			cell := fmt.Sprintf("%c2", 'A'+i)
			f.SetCellValue(sheetName, cell, data)
			f.SetCellStyle(sheetName, cell, cell, exampleStyle)
		}
	}

	return f, nil
}

