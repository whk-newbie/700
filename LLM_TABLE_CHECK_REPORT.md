# 大模型数据表与API一致性检查报告

## 检查时间
2025-01-XX

## 检查范围
- `llm_configs` 表
- `llm_call_logs` 表
- `llm_prompt_templates` 表（已废弃但保留）
- 相关模型定义
- API使用情况

---

## 1. llm_configs 表

### 当前表结构（简化后）
- `id` (SERIAL PRIMARY KEY)
- `api_key` (TEXT NOT NULL)
- `updated_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)

### 模型定义
```go
type LLMConfig struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    APIKey    string    `gorm:"type:text;not null" json:"-"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### API使用情况
- ✅ `GetOpenAIAPIKey` - 使用所有字段
- ✅ `UpdateOpenAIAPIKey` - 使用所有字段
- ✅ `RecordProxyCallLog` - 使用 `config.ID`

### 状态
✅ **完全匹配** - 表结构、模型定义和API使用完全一致

---

## 2. llm_call_logs 表

### 当前表结构
- `id` (BIGSERIAL PRIMARY KEY)
- `config_id` (INTEGER) - 无外键约束（已删除）
- `template_id` (INTEGER) - 无外键约束（已删除）
- `group_id` (INTEGER REFERENCES groups(id))
- `activation_code` (VARCHAR(32))
- `request_messages` (JSONB NOT NULL)
- `request_params` (JSONB)
- `response_content` (TEXT)
- `response_data` (JSONB)
- `status` (VARCHAR(20) NOT NULL) - CHECK (status IN ('success', 'error'))
- `error_message` (TEXT)
- `tokens_used` (INTEGER)
- `prompt_tokens` (INTEGER)
- `completion_tokens` (INTEGER)
- `call_time` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
- `duration_ms` (INTEGER)

### 模型定义
```go
type LLMCallLog struct {
    ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    ConfigID         *uint     `gorm:"type:integer" json:"config_id"`
    TemplateID       *uint     `gorm:"type:integer" json:"template_id"`
    GroupID          *uint     `gorm:"type:integer" json:"group_id"`
    ActivationCode   string    `gorm:"type:varchar(32)" json:"activation_code"`
    RequestMessages  JSONB     `gorm:"type:jsonb;not null" json:"request_messages"`
    RequestParams    JSONB     `gorm:"type:jsonb" json:"request_params"`
    ResponseContent  string    `gorm:"type:text" json:"response_content"`
    ResponseData     JSONB     `gorm:"type:jsonb" json:"response_data"`
    Status           string    `gorm:"type:varchar(20);not null" json:"status"`
    ErrorMessage     string    `gorm:"type:text" json:"error_message"`
    TokensUsed       *int      `gorm:"type:integer" json:"tokens_used"`
    PromptTokens     *int      `gorm:"type:integer" json:"prompt_tokens"`
    CompletionTokens *int      `gorm:"type:integer" json:"completion_tokens"`
    CallTime         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"call_time"`
    DurationMs       *int      `gorm:"type:integer" json:"duration_ms"`
}
```

### API使用情况
- ✅ `GetLLMCallLogs` - 查询所有字段
- ✅ `recordCallLog` - 写入所有字段
- ✅ `RecordProxyCallLog` - 写入所有字段

### 状态
✅ **完全匹配** - 表结构、模型定义和API使用完全一致

---

## 3. llm_prompt_templates 表

### 当前状态
- 表仍然存在于数据库中
- 外键约束已删除（003迁移文件）
- **不再被API使用**

### 状态
⚠️ **已废弃但保留** - 表结构保留用于历史数据兼容，但不再被使用

---

## 4. 发现的问题

### ⚠️ 问题1：迁移文件不一致

**位置**: `backend/migrations/001_init_schema.sql`

**问题描述**:
- `001_init_schema.sql` 中 `llm_call_logs` 表定义包含外键约束：
  ```sql
  config_id INTEGER REFERENCES llm_configs(id),
  template_id INTEGER REFERENCES llm_prompt_templates(id),
  ```
- 但 `003_simplify_llm_configs.sql` 删除了这些外键约束
- 对于新数据库，如果只执行 001，外键约束会存在，可能导致问题

**影响**:
- 新数据库创建时，如果只执行 001，外键约束会存在
- 如果后续执行 003，外键约束会被删除（正常）
- 但如果 003 未执行，外键约束可能导致数据插入问题

**建议**:
1. 更新 `001_init_schema.sql`，移除 `llm_call_logs` 表中的外键约束定义
2. 或者确保迁移顺序正确，003 必须执行

### ✅ 其他检查项

- ✅ 模型字段类型与数据库字段类型匹配
- ✅ 模型字段名称与数据库字段名称匹配（使用 GORM 标签）
- ✅ API 使用的字段都在表中存在
- ✅ 表约束（CHECK、NOT NULL等）与业务逻辑匹配
- ✅ 索引定义合理

---

## 5. 建议

### 立即处理
1. **更新 001_init_schema.sql**：移除 `llm_call_logs` 表中的外键约束定义，使其与当前架构一致

### 可选处理
1. **清理废弃表**：如果确定不再需要 `llm_prompt_templates` 表，可以考虑创建迁移文件删除它
2. **文档更新**：更新 DATABASE.md 文档，反映当前简化的表结构

---

## 6. 总结

### 总体状态
✅ **基本一致** - 表结构、模型定义和API使用基本匹配

### 主要问题
⚠️ **迁移文件不一致** - 001 中的外键定义与 003 的简化不一致，需要修复

### 风险等级
🟡 **低风险** - 问题只影响新数据库的创建，现有数据库通过 003 迁移已正确处理

