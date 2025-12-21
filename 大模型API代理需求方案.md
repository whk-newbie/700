# 大模型API代理 - 需求方案

> **版本**: v1.0
> **更新日期**: 2025-12-21
> **说明**: 为Windows客户端提供大模型API调用代理服务

---

## 📋 功能概述

### 核心功能
系统作为大模型API的代理，Windows客户端通过系统调用各种大模型（OpenAI、Claude、通义千问等），系统记录所有调用历史。

### 主要用途
- Windows客户端需要使用AI功能（如自动回复、内容生成等）
- 统一管理API Key
- 记录调用历史和成本
- 支持多个大模型配置

---

## 🎯 管理员配置页面

### 页面路径
`/admin/llm-config`

**权限**: ⭐ 仅管理员可访问

**说明**: 普通用户和子账号在菜单中看不到此页面

### 页面功能

#### 1. 大模型配置列表

**列表字段**:
- 配置名称
- 提供商（OpenAI、Claude、通义千问等）
- 模型名称（gpt-4、claude-3等）
- 启用状态
- 创建时间
- 操作

**操作按钮**:
- 新增配置
- 编辑
- 删除
- 启用/禁用
- 测试连接

#### 2. 新增/编辑配置对话框

**配置字段**:

##### 基础配置
- **配置名称**: 如"GPT-4配置"、"Claude配置"
- **提供商**: 下拉选择
  - OpenAI
  - Anthropic (Claude)
  - 阿里云（通义千问）
  - 讯飞星火
  - 百度文心
  - 智谱AI
  - 自定义
- **API地址**: 如 `https://api.openai.com/v1/chat/completions`
- **API Key**: 加密存储
- **模型名称**: 如 `gpt-4`、`claude-3-opus`

##### 参数配置
- **Max Tokens**: 滑块或输入框（1-32000）
- **Temperature**: 滑块（0.0-2.0）
- **Top P**: 滑块（0.0-1.0）
- **Frequency Penalty**: 滑块（-2.0-2.0）
- **Presence Penalty**: 滑块（-2.0-2.0）

##### Prompt配置
- **System Prompt**: 多行文本输入框
- **Prompt模板**: 支持多个模板配置
  - 模板名称
  - 模板内容
  - 变量占位符（如 {customer_name}、{message}）

**示例**:
```
模板1：客户问候
---
你是一个专业的客服人员，客户名叫{customer_name}。
请用友好的语气回复客户的问题：{message}

模板2：内容总结
---
请总结以下对话内容：
{conversation}
```

##### 高级配置
- **超时时间**: 秒（默认30）
- **重试次数**: 次（默认3）
- **是否启用**: 开关

---

## 🔌 Windows客户端调用接口

### 接口说明

**目的**: Windows客户端调用大模型API进行AI处理

**认证**: 使用激活码对应的JWT Token

### API接口

#### 1. 获取可用配置列表

**接口**: `GET /api/llm/configs`

**请求头**:
```
Authorization: Bearer {jwt_token}
```

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "GPT-4配置",
      "provider": "openai",
      "model": "gpt-4",
      "is_active": true
    },
    {
      "id": 2,
      "name": "Claude配置",
      "provider": "anthropic",
      "model": "claude-3-opus",
      "is_active": true
    }
  ]
}
```

**说明**: 只返回已启用的配置

---

#### 2. 调用大模型API

**接口**: `POST /api/llm/call`

**请求头**:
```
Authorization: Bearer {jwt_token}
```

**请求体**:
```json
{
  "config_id": 1,                      // 使用哪个配置
  "messages": [
    {
      "role": "user",
      "content": "你好，请帮我生成一段欢迎语"
    }
  ],
  "template_id": 1,                    // 可选：使用哪个Prompt模板
  "template_variables": {              // 可选：模板变量
    "customer_name": "张三",
    "message": "你好"
  },
  "stream": false                      // 是否流式返回
}
```

**响应（非流式）**:
```json
{
  "success": true,
  "data": {
    "response": "你好！很高兴为您服务...",
    "model": "gpt-4",
    "tokens_used": 150,
    "call_log_id": 12345
  }
}
```

**响应（流式）**:
```
data: {"type":"chunk","content":"你好"}
data: {"type":"chunk","content":"！"}
data: {"type":"chunk","content":"很高兴"}
data: {"type":"done","tokens_used":150}
```

---

#### 3. 获取Prompt模板列表

**接口**: `GET /api/llm/templates?config_id=1`

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "客户问候",
      "content": "你是一个专业的客服人员...",
      "variables": ["customer_name", "message"]
    },
    {
      "id": 2,
      "name": "内容总结",
      "content": "请总结以下对话内容：{conversation}",
      "variables": ["conversation"]
    }
  ]
}
```

---

#### 4. 查看调用历史

**接口**: `GET /api/llm/logs`

**查询参数**:
- `config_id`: 配置ID（可选）
- `start_date`: 开始日期
- `end_date`: 结束日期
- `status`: success / error
- `page`: 页码
- `page_size`: 每页数量

**响应**:
```json
{
  "success": true,
  "total": 1000,
  "page": 1,
  "page_size": 20,
  "data": [
    {
      "id": 12345,
      "config_name": "GPT-4配置",
      "model": "gpt-4",
      "request_preview": "你好，请帮我...",
      "response_preview": "你好！很高兴...",
      "tokens_used": 150,
      "status": "success",
      "duration_ms": 1200,
      "created_at": "2025-12-21 10:30:00"
    }
  ]
}
```

---

## 💾 数据库表设计

### llm_configs（大模型配置表）

```sql
CREATE TABLE llm_configs (
    id SERIAL PRIMARY KEY,
    
    -- 基本信息
    name VARCHAR(100) NOT NULL,                        -- 配置名称
    provider VARCHAR(50) NOT NULL,                     -- 提供商
    
    -- API配置
    api_url VARCHAR(500) NOT NULL,                     -- API地址
    api_key TEXT NOT NULL,                             -- API Key（加密存储）
    model VARCHAR(100) NOT NULL,                       -- 模型名称
    
    -- 参数配置
    max_tokens INTEGER DEFAULT 2000,
    temperature DECIMAL(3,2) DEFAULT 0.7,
    top_p DECIMAL(3,2) DEFAULT 1.0,
    frequency_penalty DECIMAL(3,2) DEFAULT 0.0,
    presence_penalty DECIMAL(3,2) DEFAULT 0.0,
    
    -- System Prompt
    system_prompt TEXT,                                -- 系统提示词
    
    -- 高级配置
    timeout_seconds INTEGER DEFAULT 30,                -- 超时时间
    max_retries INTEGER DEFAULT 3,                     -- 重试次数
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    
    -- 约束
    CONSTRAINT check_provider CHECK (provider IN (
        'openai', 'anthropic', 'aliyun', 'xunfei', 'baidu', 'zhipu', 'custom'
    ))
);

-- 索引
CREATE INDEX idx_llm_configs_provider ON llm_configs(provider);
CREATE INDEX idx_llm_configs_is_active ON llm_configs(is_active);

-- 注释
COMMENT ON TABLE llm_configs IS '大模型配置表';
COMMENT ON COLUMN llm_configs.api_key IS 'API Key（使用AES加密存储）';
```

---

### llm_prompt_templates（Prompt模板表）

```sql
CREATE TABLE llm_prompt_templates (
    id SERIAL PRIMARY KEY,
    
    -- 关联配置
    config_id INTEGER NOT NULL REFERENCES llm_configs(id) ON DELETE CASCADE,
    
    -- 模板信息
    template_name VARCHAR(100) NOT NULL,               -- 模板名称
    template_content TEXT NOT NULL,                    -- 模板内容
    
    -- 变量定义
    variables JSONB,                                   -- 变量列表 ["customer_name", "message"]
    
    -- 描述
    description TEXT,                                  -- 模板说明
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_prompt_templates_config ON llm_prompt_templates(config_id);
CREATE INDEX idx_prompt_templates_active ON llm_prompt_templates(is_active);

-- 注释
COMMENT ON TABLE llm_prompt_templates IS 'Prompt模板表';
COMMENT ON COLUMN llm_prompt_templates.variables IS '模板变量列表（JSON数组）';
```

---

### llm_call_logs（大模型调用记录表）

```sql
CREATE TABLE llm_call_logs (
    id BIGSERIAL PRIMARY KEY,
    
    -- 关联信息
    config_id INTEGER REFERENCES llm_configs(id),
    template_id INTEGER REFERENCES llm_prompt_templates(id),
    
    -- 调用者信息
    group_id INTEGER REFERENCES groups(id),            -- 哪个分组调用的
    activation_code VARCHAR(32),                       -- 激活码
    
    -- 请求信息
    request_messages JSONB NOT NULL,                   -- 请求消息（完整JSON）
    request_params JSONB,                              -- 请求参数
    
    -- 响应信息
    response_content TEXT,                             -- 响应内容
    response_data JSONB,                               -- 完整响应（JSON）
    
    -- 状态信息
    status VARCHAR(20) NOT NULL,                       -- 'success' | 'error'
    error_message TEXT,                                -- 错误信息
    
    -- 统计信息
    tokens_used INTEGER,                               -- Token使用量
    prompt_tokens INTEGER,                             -- Prompt Token
    completion_tokens INTEGER,                         -- Completion Token
    
    -- 时间信息
    call_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,     -- 调用时间
    duration_ms INTEGER,                               -- 耗时（毫秒）
    
    -- 约束
    CONSTRAINT check_status CHECK (status IN ('success', 'error'))
);

-- 索引
CREATE INDEX idx_llm_call_logs_config ON llm_call_logs(config_id, call_time DESC);
CREATE INDEX idx_llm_call_logs_group ON llm_call_logs(group_id, call_time DESC);
CREATE INDEX idx_llm_call_logs_activation ON llm_call_logs(activation_code);
CREATE INDEX idx_llm_call_logs_time ON llm_call_logs(call_time DESC);
CREATE INDEX idx_llm_call_logs_status ON llm_call_logs(status);

-- 注释
COMMENT ON TABLE llm_call_logs IS '大模型调用记录表';
COMMENT ON COLUMN llm_call_logs.tokens_used IS 'Token总使用量';
```

---

## 🔐 API Key加密存储

### 加密方案

**使用AES-256加密**:

```go
// Go代码示例
import "crypto/aes"
import "crypto/cipher"

// 加密API Key
func EncryptAPIKey(plaintext, secretKey string) (string, error) {
    block, err := aes.NewCipher([]byte(secretKey))
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    // ... 生成随机nonce
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密API Key
func DecryptAPIKey(ciphertext, secretKey string) (string, error) {
    // 解密逻辑
    // ...
}
```

**环境变量**:
```bash
# .env文件
ENCRYPTION_SECRET_KEY=your-32-byte-secret-key-here
```

---

## 🔌 API接口设计

### 管理员接口（配置管理）

#### 1. 获取配置列表

**接口**: `GET /api/admin/llm/configs`

**权限**: 仅管理员

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "GPT-4配置",
      "provider": "openai",
      "model": "gpt-4",
      "api_url": "https://api.openai.com/v1/chat/completions",
      "max_tokens": 2000,
      "temperature": 0.7,
      "is_active": true,
      "created_at": "2025-12-21 10:00:00"
    }
  ]
}
```

**说明**: API Key不返回给前端（安全）

---

#### 2. 创建配置

**接口**: `POST /api/admin/llm/configs`

**请求**:
```json
{
  "name": "GPT-4配置",
  "provider": "openai",
  "api_url": "https://api.openai.com/v1/chat/completions",
  "api_key": "sk-xxxxxxxxxxxxxxxx",
  "model": "gpt-4",
  "max_tokens": 2000,
  "temperature": 0.7,
  "top_p": 1.0,
  "system_prompt": "你是一个专业的助手..."
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "GPT-4配置"
  }
}
```

---

#### 3. 更新配置

**接口**: `PUT /api/admin/llm/configs/:id`

#### 4. 删除配置

**接口**: `DELETE /api/admin/llm/configs/:id`

#### 5. 测试配置

**接口**: `POST /api/admin/llm/configs/:id/test`

**功能**: 测试API Key是否有效

**请求**:
```json
{
  "test_message": "Hello"
}
```

**响应**:
```json
{
  "success": true,
  "message": "配置测试成功",
  "response": "Hello! How can I help you?",
  "duration_ms": 1200
}
```

---

### Prompt模板管理接口

#### 1. 获取模板列表

**接口**: `GET /api/admin/llm/templates?config_id=1`

#### 2. 创建模板

**接口**: `POST /api/admin/llm/templates`

**请求**:
```json
{
  "config_id": 1,
  "template_name": "客户问候",
  "template_content": "你是一个专业的客服，客户名叫{customer_name}。请回复：{message}",
  "variables": ["customer_name", "message"],
  "description": "用于客户问候的模板"
}
```

#### 3. 更新模板

**接口**: `PUT /api/admin/llm/templates/:id`

#### 4. 删除模板

**接口**: `DELETE /api/admin/llm/templates/:id`

---

### 客户端调用接口（核心）

#### 1. 调用大模型（标准模式）

**接口**: `POST /api/llm/call`

**权限**: Windows客户端（需JWT Token）

**请求**:
```json
{
  "config_id": 1,
  "messages": [
    {
      "role": "system",
      "content": "你是一个助手"
    },
    {
      "role": "user",
      "content": "你好"
    }
  ],
  "max_tokens": 1000,                  // 可选，覆盖配置
  "temperature": 0.8                   // 可选，覆盖配置
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "response": "你好！有什么可以帮助你的吗？",
    "model": "gpt-4",
    "tokens_used": 50,
    "prompt_tokens": 20,
    "completion_tokens": 30,
    "call_log_id": 12345,
    "duration_ms": 1200
  }
}
```

---

#### 2. 使用模板调用

**接口**: `POST /api/llm/call-template`

**请求**:
```json
{
  "template_id": 1,
  "variables": {
    "customer_name": "张三",
    "message": "你好"
  }
}
```

**服务器处理**:
1. 获取模板内容
2. 替换变量：`你是一个专业的客服，客户名叫张三。请回复：你好`
3. 调用大模型API
4. 返回结果

---

#### 3. 流式调用

**接口**: `POST /api/llm/call-stream`

**响应**: Server-Sent Events (SSE)

```
event: chunk
data: {"content":"你好"}

event: chunk
data: {"content":"！"}

event: done
data: {"tokens_used":50}
```

---

## 📊 调用记录查询页面

### 页面路径
`/admin/llm-logs`

**权限**: ⭐ 仅管理员可访问

**说明**: 
- 管理员可以查看所有分组的调用记录
- 普通用户和子账号无法访问此页面
- 在菜单中对普通用户和子账号隐藏

### 页面功能

#### 筛选条件
- 时间范围
- 配置选择（下拉）
- 状态（成功/失败）
- 激活码（管理员可见）

#### 列表字段
- 调用ID
- 配置名称
- 模型
- 请求预览（前50字）
- 响应预览（前50字）
- Token使用量
- 耗时（ms）
- 状态
- 调用时间
- 操作（查看详情）

#### 查看详情对话框

**显示内容**:
- 完整请求内容
- 完整响应内容
- Token统计
- 耗时统计
- 错误信息（如果失败）

---

## 🔄 大模型调用流程

### 完整流程

```
1. Windows客户端需要调用AI功能
   ↓
2. 调用服务器API: POST /api/llm/call
   请求头携带JWT Token（激活码认证）
   ↓
3. 服务器验证Token，获取group_id和activation_code
   ↓
4. 获取大模型配置
   - 根据config_id获取配置
   - 解密API Key
   ↓
5. 构造请求
   - 如果使用模板，替换变量
   - 合并System Prompt
   - 应用参数配置
   ↓
6. 调用大模型API
   - 使用配置的api_url和api_key
   - 设置超时和重试
   ↓
7. 接收响应
   ↓
8. 记录调用日志
   - 插入llm_call_logs表
   - 记录请求、响应、Token、耗时
   ↓
9. 返回结果给客户端
```

---

## 🔒 安全设计

### 1. API Key保护
- **存储**: AES-256加密存储
- **传输**: HTTPS加密传输
- **访问**: 只有后端能解密，前端永远看不到
- **日志**: API Key不记录在日志中

### 2. 权限控制
- **管理员**: 管理所有配置、查看所有调用记录
- **普通用户**: ❌ 无法查看配置和调用记录
- **子账号**: ❌ 无法查看配置和调用记录
- **Windows客户端**: 只能调用API，但调用者看不到配置和日志

**说明**: 
- 只有管理员能看到大模型相关的所有信息
- 普通用户和子账号虽然可以通过Windows客户端调用大模型，但在后台看不到任何大模型配置和调用记录
- 所有调用记录都记录在数据库中，但只有管理员能查看

### 3. 调用限制（可选）
- 每个分组的调用频率限制
- 每日Token使用量限制
- 成本控制

---

## 📈 统计与监控

### 调用统计

**按配置统计**:
- 总调用次数
- 成功率
- 平均耗时
- Token总消耗

**按分组统计**:
- 各分组的调用次数
- Token消耗排行
- 成本分析

**按时间统计**:
- 每日调用趋势
- 每小时调用分布
- Token消耗趋势

---

## 🎨 前端页面设计

### 大模型配置页面（管理员）

**页面布局**:
```
┌─────────────────────────────────────────────────────────┐
│  🤖 大模型配置                                          │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  [➕ 新增配置]  [📊 调用统计]                            │
│                                                           │
│  ┌──┬────────┬────────┬──────┬────┬────────┬────┐      │
│  │ID│配置名称│提供商  │模型  │状态│创建时间│操作│      │
│  ├──┼────────┼────────┼──────┼────┼────────┼────┤      │
│  │1 │GPT-4   │OpenAI  │gpt-4 │启用│12-21   │编辑│      │
│  │2 │Claude  │Anthro  │claude│启用│12-20   │编辑│      │
│  └──┴────────┴────────┴──────┴────┴────────┴────┘      │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### 配置编辑对话框

**布局**:
```
┌─────────────────────────────────────┐
│  编辑大模型配置            [关闭 ✕] │
├─────────────────────────────────────┤
│                                       │
│  基础配置                             │
│  配置名称：[GPT-4配置          ]     │
│  提供商：  [OpenAI ▼]               │
│  API地址： [https://api...    ]     │
│  API Key： [**********    ] [显示]  │
│  模型：    [gpt-4             ]     │
│                                       │
│  参数配置                             │
│  Max Tokens：    [====●====] 2000   │
│  Temperature：   [===●=====] 0.7    │
│  Top P：         [=========●] 1.0   │
│                                       │
│  Prompt配置                          │
│  System Prompt：                     │
│  ┌─────────────────────────────┐   │
│  │你是一个专业的助手...        │   │
│  └─────────────────────────────┘   │
│                                       │
│  [管理Prompt模板]                    │
│                                       │
│  高级配置                             │
│  超时时间：[30] 秒                   │
│  重试次数：[3] 次                    │
│  启用：☑                             │
│                                       │
│         [测试连接] [保存] [取消]     │
│                                       │
└─────────────────────────────────────┘
```

### Prompt模板管理对话框

```
┌─────────────────────────────────────┐
│  Prompt模板管理            [关闭 ✕] │
├─────────────────────────────────────┤
│                                       │
│  [➕ 新增模板]                       │
│                                       │
│  ┌──┬────────┬────────┬────┬────┐  │
│  │ID│模板名称│变量    │状态│操作│  │
│  ├──┼────────┼────────┼────┼────┤  │
│  │1 │客户问候│name,msg│启用│编辑│  │
│  │2 │内容总结│conv    │启用│编辑│  │
│  └──┴────────┴────────┴────┴────┘  │
│                                       │
└─────────────────────────────────────┘
```

---

## 📊 调用记录页面

### 页面布局

```
┌─────────────────────────────────────────────────────────┐
│  📋 大模型调用记录                                       │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  [筛选区域]                                               │
│  ┌──────────┬──────────┬──────────┬──────────┐          │
│  │开始时间～│ 配置     │ 状态     │ 激活码   │          │
│  │截止时间  │          │          │(管理员)  │          │
│  └──────────┴──────────┴──────────┴──────────┘          │
│                                                           │
│  [🔍 搜索]  [🔄 重选]                                    │
│                                                           │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  统计卡片：                                               │
│  ┌──────────┬──────────┬──────────┬──────────┐          │
│  │总调用次数│成功率    │平均耗时  │Token消耗│          │
│  │  1,234   │  98.5%   │  1.2s    │ 125K    │          │
│  └──────────┴──────────┴──────────┴──────────┘          │
│                                                           │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  [数据表格]                                               │
│  ┌──┬────┬────┬────────┬────────┬─────┬────┬────┬────┐ │
│  │ID│配置│模型│请求预览│响应预览│Token│耗时│状态│操作│ │
│  ├──┼────┼────┼────────┼────────┼─────┼────┼────┼────┤ │
│  │1 │GPT4│gpt │你好... │你好！..│150  │1.2s│成功│详情│ │
│  │2 │Claude│claude│总结..│以下是..│200  │0.9s│成功│详情│ │
│  └──┴────┴────┴────────┴────────┴─────┴────┴────┴────┘ │
│                                                           │
│  共 1234 条  [20条/页▼] [◀][1][2][3][4][▶]             │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### 详情对话框

```
┌─────────────────────────────────────┐
│  调用详情                  [关闭 ✕] │
├─────────────────────────────────────┤
│                                       │
│  基本信息：                           │
│  - 配置：GPT-4配置                   │
│  - 模型：gpt-4                       │
│  - 激活码：ABC123                    │
│  - 调用时间：2025-12-21 10:30:00    │
│  - 耗时：1200ms                      │
│  - 状态：成功                        │
│                                       │
│  Token统计：                         │
│  - Prompt Token：80                  │
│  - Completion Token：70              │
│  - 总计：150                         │
│                                       │
│  请求内容：                           │
│  ┌─────────────────────────────┐   │
│  │[                             │   │
│  │  {                           │   │
│  │    "role": "user",           │   │
│  │    "content": "你好"         │   │
│  │  }                           │   │
│  │]                             │   │
│  └─────────────────────────────┘   │
│                                       │
│  响应内容：                           │
│  ┌─────────────────────────────┐   │
│  │你好！很高兴为您服务。有什么│   │
│  │可以帮助您的吗？              │   │
│  └─────────────────────────────┘   │
│                                       │
│              [关闭]                  │
│                                       │
└─────────────────────────────────────┘
```

---

## 🔧 后端实现要点

### Go代码结构示例

```go
// internal/services/llm_service.go

type LLMService struct {
    db    *gorm.DB
    redis *redis.Client
}

// 调用大模型
func (s *LLMService) CallLLM(req *CallRequest, groupID uint, activationCode string) (*CallResponse, error) {
    // 1. 获取配置
    var config models.LLMConfig
    if err := s.db.First(&config, req.ConfigID).Error; err != nil {
        return nil, err
    }
    
    // 2. 解密API Key
    apiKey, err := DecryptAPIKey(config.APIKey)
    if err != nil {
        return nil, err
    }
    
    // 3. 构造请求
    messages := req.Messages
    if config.SystemPrompt != "" {
        messages = append([]Message{{Role: "system", Content: config.SystemPrompt}}, messages...)
    }
    
    // 4. 调用大模型API
    startTime := time.Now()
    response, err := s.callProvider(config, apiKey, messages, req)
    duration := time.Since(startTime)
    
    // 5. 记录日志
    log := &models.LLMCallLog{
        ConfigID:         config.ID,
        GroupID:          groupID,
        ActivationCode:   activationCode,
        RequestMessages:  messages,
        ResponseContent:  response.Content,
        Status:           ifErr(err, "error", "success"),
        TokensUsed:       response.TokensUsed,
        PromptTokens:     response.PromptTokens,
        CompletionTokens: response.CompletionTokens,
        DurationMS:       int(duration.Milliseconds()),
    }
    s.db.Create(log)
    
    return response, err
}

// 调用具体提供商
func (s *LLMService) callProvider(config models.LLMConfig, apiKey string, messages []Message, req *CallRequest) (*Response, error) {
    switch config.Provider {
    case "openai":
        return s.callOpenAI(config, apiKey, messages, req)
    case "anthropic":
        return s.callAnthropic(config, apiKey, messages, req)
    // ... 其他提供商
    default:
        return s.callCustom(config, apiKey, messages, req)
    }
}
```

---

## 🎯 支持的大模型提供商

### 1. OpenAI
- **模型**: gpt-4, gpt-3.5-turbo, gpt-4-turbo
- **API**: https://api.openai.com/v1/chat/completions

### 2. Anthropic (Claude)
- **模型**: claude-3-opus, claude-3-sonnet, claude-3-haiku
- **API**: https://api.anthropic.com/v1/messages

### 3. 阿里云（通义千问）
- **模型**: qwen-turbo, qwen-plus, qwen-max
- **API**: https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation

### 4. 讯飞星火
- **模型**: spark-3.5, spark-3.0
- **API**: https://spark-api.xf-yun.com/v1/chat/completions

### 5. 百度文心
- **模型**: ernie-bot-4, ernie-bot-turbo
- **API**: https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions

### 6. 智谱AI
- **模型**: chatglm_turbo, chatglm_pro
- **API**: https://open.bigmodel.cn/api/paas/v4/chat/completions

### 7. 自定义
- **支持**: 任何兼容OpenAI格式的API

---

## 💰 成本控制（可选）

### 配额管理

**为每个分组设置配额**:
```sql
ALTER TABLE groups ADD COLUMN daily_token_limit INTEGER DEFAULT NULL;
ALTER TABLE groups ADD COLUMN monthly_token_limit INTEGER DEFAULT NULL;
```

**检查逻辑**:
```go
func (s *LLMService) CheckQuota(groupID uint) error {
    // 查询今日已使用Token
    var todayUsed int
    s.db.Model(&LLMCallLog{}).
        Where("group_id = ? AND call_time >= ?", groupID, todayStart).
        Select("COALESCE(SUM(tokens_used), 0)").
        Scan(&todayUsed)
    
    // 检查是否超限
    if group.DailyTokenLimit != nil && todayUsed >= *group.DailyTokenLimit {
        return errors.New("今日Token配额已用完")
    }
    
    return nil
}
```

---

## ✅ 已确认需求

- [x] 大模型配置管理（管理员）
- [x] 支持多个大模型提供商
- [x] Prompt模板管理
- [x] Windows客户端调用接口
- [x] 调用记录查询
- [x] API Key加密存储
- [x] 调用统计与监控
- [x] 数据库表设计
- [x] 权限控制
- [x] 流式调用支持（可选）
- [x] 成本控制（可选）


