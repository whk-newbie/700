# Line账号管理系统 - API文档规划

> **版本**: v1.0
> **更新日期**: 2025-12-21
> **说明**: API接口文档化规范和标准

---

## 🎯 文档目标

### 主要目标
- **标准化**: 统一API接口文档格式和规范
- **自动化**: 支持Swagger/OpenAPI自动生成
- **可维护**: 文档与代码同步更新
- **易阅读**: 清晰的接口说明和示例

### 受众
- **前端开发者**: 了解接口调用方式
- **后端开发者**: 接口实现和维护
- **测试人员**: 接口测试用例编写
- **产品经理**: 功能验证和需求确认
- **第三方集成**: 外部系统对接

---

## 📋 API文档体系

### 1. Swagger/OpenAPI 3.0规范

#### 全局配置
```yaml
openapi: 3.0.3
info:
  title: Line账号管理系统API
  description: Line账号分组管理与进线统计系统API
  version: v1.0.0
  contact:
    name: API Support
    email: support@line-management.com

servers:
  - url: http://localhost:8080/api/v1
    description: 开发环境
  - url: https://api.line-management.com/api/v1
    description: 生产环境

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    # 通用响应格式
    ApiResponse:
      type: object
      properties:
        code:
          type: integer
          description: 业务状态码
          example: 1000
        message:
          type: string
          description: 响应消息
          example: "操作成功"
        data:
          description: 响应数据
        timestamp:
          type: string
          format: date-time
          description: 响应时间
          example: "2025-12-21T10:00:00Z"

    # 分页响应格式
    PaginatedResponse:
      allOf:
        - $ref: '#/components/schemas/ApiResponse'
        - type: object
          properties:
            data:
              type: object
              properties:
                list:
                  type: array
                  description: 数据列表
                pagination:
                  type: object
                  properties:
                    page:
                      type: integer
                      example: 1
                    page_size:
                      type: integer
                      example: 10
                    total:
                      type: integer
                      example: 100
                    total_pages:
                      type: integer
                      example: 10
```

### 2. 接口分组和版本控制

#### API版本策略
```
/api/v1/          # v1版本（当前）
/api/v2/          # v2版本（未来扩展）
```

#### 接口分组
```
/api/v1/auth/           # 认证授权
/api/v1/groups/         # 分组管理
/api/v1/line-accounts/  # Line账号管理
/api/v1/customers/      # 客户管理
/api/v1/follow-ups/     # 跟进记录
/api/v1/stats/          # 统计数据
/api/v1/contact-pool/   # 底库管理
/api/v1/admin/          # 管理员功能
/api/v1/llm/            # 大模型服务
/api/v1/ws/             # WebSocket服务
```

---

## 🔧 接口文档规范

### 1. 接口基本信息

#### 必填字段
```yaml
paths:
  /api/v1/groups:
    get:
      tags:
        - 分组管理
      summary: 获取分组列表
      description: |
        获取当前用户有权限的分组列表，支持分页和筛选

        **权限要求**: Admin, User, SubAccount
        **数据过滤**: SubAccount只能看到自己分组的数据
      operationId: getGroups
      security:
        - BearerAuth: []
```

#### 参数定义
```yaml
parameters:
  - name: page
    in: query
    description: 页码（从1开始）
    required: false
    schema:
      type: integer
      minimum: 1
      default: 1
      example: 1

  - name: page_size
    in: query
    description: 每页数量
    required: false
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 10
      example: 10

  - name: keyword
    in: query
    description: 搜索关键词（分组名称、激活码）
    required: false
    schema:
      type: string
      maxLength: 50
      example: "测试分组"

  - name: status
    in: query
    description: 状态筛选
    required: false
    schema:
      type: string
      enum: [active, disabled]
      example: "active"
```

### 2. 请求体和响应体

#### 请求体示例
```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        type: object
        required:
          - name
          - category
        properties:
          name:
            type: string
            description: 分组名称
            minLength: 1
            maxLength: 50
            example: "销售一部"
          category:
            type: string
            description: 分组分类
            enum: [sales, support, marketing]
            example: "sales"
          description:
            type: string
            description: 分组描述
            maxLength: 200
            example: "负责一线销售工作"
```

#### 响应体示例
```yaml
responses:
  '200':
    description: 成功响应
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/ApiResponse'
            - type: object
              properties:
                data:
                  type: object
                  properties:
                    id:
                      type: integer
                      example: 1
                    name:
                      type: string
                      example: "销售一部"
                    activation_code:
                      type: string
                      example: "ABC123456"
                    status:
                      type: string
                      enum: [active, disabled]
                      example: "active"
                    created_at:
                      type: string
                      format: date-time
                      example: "2025-12-21T08:00:00Z"
                    updated_at:
                      type: string
                      format: date-time
                      example: "2025-12-21T10:00:00Z"
```

### 3. 错误响应

#### 标准错误响应
```yaml
responses:
  '400':
    description: 参数错误
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/ApiResponse'
            - type: object
              properties:
                code:
                  example: 1001
                message:
                  example: "分组名称不能为空"

  '401':
    description: 未授权
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/ApiResponse'
            - type: object
              properties:
                code:
                  example: 2001
                message:
                  example: "请先登录"

  '403':
    description: 权限不足
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/ApiResponse'
            - type: object
              properties:
                code:
                  example: 2007
                message:
                  example: "权限不足"

  '404':
    description: 资源不存在
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/ApiResponse'
            - type: object
              properties:
                code:
                  example: 3001
                message:
                  example: "分组不存在"
```

---

## 📚 接口文档模板

### 1. 接口概述模板

#### 接口信息卡片
```
## 接口: 获取分组列表

**接口地址**: `GET /api/v1/groups`

**权限要求**: Admin, User, SubAccount

**数据过滤**: SubAccount只能看到自己分组的数据

**接口描述**:
获取当前用户有权限的分组列表，支持分页、搜索和状态筛选。
```

#### 参数表格
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | integer | 否 | 1 | 页码（从1开始） |
| page_size | integer | 否 | 10 | 每页数量（1-100） |
| keyword | string | 否 | - | 搜索关键词 |
| status | string | 否 | - | 状态筛选：active/disabled |

#### 请求示例
```bash
GET /api/v1/groups?page=1&page_size=10&keyword=销售&status=active
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 响应示例
```json
{
  "code": 1000,
  "message": "查询成功",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "销售一部",
        "activation_code": "ABC123456",
        "status": "active",
        "category": "sales",
        "description": "负责一线销售",
        "created_at": "2025-12-21T08:00:00Z",
        "updated_at": "2025-12-21T10:00:00Z",
        "stats": {
          "total_accounts": 5,
          "online_accounts": 3,
          "today_incoming": 25
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 1,
      "total_pages": 1
    }
  },
  "timestamp": "2025-12-21T10:00:00Z"
}
```

### 2. 业务逻辑说明

#### 权限控制说明
- **Admin**: 可以查看所有分组
- **User**: 可以查看所有分组（普通用户）
- **SubAccount**: 只能查看自己所属的分组

#### 数据过滤规则
- SubAccount用户的数据会自动按激活码过滤
- 统计数据包含实时在线账号数和今日进线数

#### 错误情况
- 无效的分页参数：返回参数错误
- 数据库查询失败：返回服务器错误

---

## 🛠️ 文档生成工具

### 1. Go Swagger注解

#### 基本注解
```go
// @title Line账号管理系统API
// @version 1.0
// @description Line账号分组管理与进线统计系统API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token，格式：Bearer {token}

package main

// @Summary 获取分组列表
// @Description 获取当前用户有权限的分组列表
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param status query string false "状态筛选" Enums(active,disabled)
// @Success 200 {object} schemas.GroupListResponse
// @Failure 400 {object} schemas.ErrorResponse
// @Failure 401 {object} schemas.ErrorResponse
// @Failure 500 {object} schemas.ErrorResponse
// @Router /groups [get]
// @Security BearerAuth
func GetGroups(c *gin.Context) {
    // 实现逻辑
}
```

### 2. 文档生成命令

```bash
# 安装swaggo工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成API文档
swag init -g cmd/server/main.go -o docs/

# 启动服务后访问文档
# http://localhost:8080/swagger/index.html
```

### 3. 文档验证

```bash
# 使用swagger-codegen验证API规范
swagger-codegen validate -i docs/swagger.json

# 生成客户端代码（可选）
swagger-codegen generate -i docs/swagger.json -l javascript -o client/
```

---

## 📋 文档维护流程

### 1. 接口开发流程
1. **需求确认**: 与产品经理确认接口需求
2. **接口设计**: 设计请求/响应格式，确定权限要求
3. **编写注解**: 在Go代码中添加Swagger注解
4. **实现接口**: 编写业务逻辑
5. **文档生成**: 执行swag init生成文档
6. **测试验证**: 前端联调测试
7. **文档更新**: 根据测试结果更新文档

### 2. 文档审查清单

#### 功能完整性
- [ ] 接口地址正确
- [ ] 请求方法正确
- [ ] 参数定义完整
- [ ] 响应格式正确
- [ ] 错误处理完整

#### 文档质量
- [ ] 描述清晰准确
- [ ] 示例数据合理
- [ ] 参数说明详细
- [ ] 权限要求明确
- [ ] 业务逻辑说明

#### 规范一致性
- [ ] 命名规范统一
- [ ] 数据格式一致
- [ ] 状态码使用正确
- [ ] 注解格式标准

### 3. 版本管理

#### 文档版本控制
```
docs/
├── v1.0/              # v1.0版本文档
│   ├── swagger.json
│   ├── swagger.yaml
│   └── index.html
├── v1.1/              # v1.1版本文档
└── latest -> v1.1/    # 最新版本软链接
```

#### 版本更新说明
- **主版本号**: 不兼容的API变更
- **次版本号**: 向后兼容的功能新增
- **修订号**: 向后兼容的问题修复

---

## 🔍 文档质量保证

### 1. 自动化检查

#### API规范检查脚本
```bash
#!/bin/bash
# api_check.sh - API文档质量检查

echo "🔍 检查API文档质量..."

# 1. 验证Swagger格式
swagger-codegen validate -i docs/swagger.json

# 2. 检查必填字段
jq -e '.info.title' docs/swagger.json > /dev/null || echo "❌ 缺少title"
jq -e '.info.version' docs/swagger.json > /dev/null || echo "❌ 缺少version"

# 3. 检查接口覆盖率
total_endpoints=$(jq '.paths | length' docs/swagger.json)
echo "📊 接口总数: $total_endpoints"

# 4. 检查是否有未文档化的接口
grep -r "@Router" internal/handlers/ | wc -l
```

### 2. 文档测试

#### 接口测试覆盖
- **单元测试**: 业务逻辑测试
- **集成测试**: 接口联调测试
- **文档测试**: 基于Swagger的自动化测试

#### 文档准确性验证
```javascript
// 使用swagger-js测试文档准确性
const Swagger = require('swagger-client');

Swagger('http://localhost:8080/swagger.json')
  .then(client => {
    return client.apis.groups.getGroups();
  })
  .then(response => {
    console.log('✅ 接口调用成功');
  })
  .catch(error => {
    console.log('❌ 接口调用失败:', error);
  });
```

---

## 📖 相关文档

- [状态码定义.md](状态码定义.md) - 统一的错误状态码定义
- [项目实施规划.md](项目实施规划.md) - 项目开发计划
- [数据库表设计-完整版.md](数据库表设计-完整版.md) - 数据库设计文档
- [Windows客户端交互协议.md](Windows客户端交互协议.md) - 客户端对接协议

---

## 🎯 实施计划

### 第14周：部署与文档
- [x] API文档规划（本周完成）
- [ ] Swagger集成到Go项目
- [ ] 编写接口注解
- [ ] 生成API文档
- [ ] 文档部署和访问配置

### 后续维护
- [ ] 接口变更时同步更新文档
- [ ] 定期审查文档质量
- [ ] 根据用户反馈优化文档
- [ ] 维护API变更日志