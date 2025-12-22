# API 文档查看指南

## 📖 如何查看 API 文档

### 方法一：Swagger UI（推荐，最简单）

1. **启动后端服务**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. **打开浏览器访问 Swagger UI**
   - 访问地址：`http://localhost:8080/swagger/index.html`
   - 或者：`http://localhost:8080/swagger/doc.json` 查看 JSON 格式

3. **在 Swagger UI 中测试 API**
   - 点击右上角的 **"Authorize"** 按钮 🔒
   - 输入 JWT Token（格式：`Bearer {your_token}`）
   - 例如：`Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
   - 点击 "Authorize" 确认
   - 现在可以在页面上直接测试各个 API 接口了！

### 方法二：查看生成的文档文件

文档已生成在 `backend/docs/` 目录：
- `swagger.json` - JSON 格式的 API 文档
- `swagger.yaml` - YAML 格式的 API 文档
- `docs.go` - Go 代码文件（用于嵌入文档）

## 🔄 更新 API 文档

当你添加或修改了 API 接口后，需要重新生成文档：

```bash
cd backend
swag init -g cmd/server/main.go -o docs
```

**注意**：如果 `swag` 命令找不到，需要先安装：
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

**最新更新**：
- 新增接口：`POST /api/v1/groups/:id/generate-subaccount-token` - 生成子账户Token
  - 支持管理员和普通用户使用
  - 管理员可为任何分组生成Token
  - 普通用户只能为自己管理的分组生成Token
  - 生成的Token可用于在新标签页自动登录子账户界面

## 🔐 获取 Token 进行测试

1. **登录获取 Token**
   - 使用 `POST /api/v1/auth/login` 接口
   - 请求体：
     ```json
     {
       "username": "admin",
       "password": "your_password"
     }
     ```
   - 响应中会返回 `token` 字段

2. **在 Swagger UI 中使用 Token**
   - 复制返回的 token
   - 在 Swagger UI 的 "Authorize" 对话框中输入：`Bearer {token}`
   - 点击 "Authorize" 确认

## 📋 当前可用的 API 接口

### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/login-subaccount` - 子账号登录
- `POST /api/v1/auth/logout` - 登出
- `GET /api/v1/auth/me` - 获取当前用户信息
- `POST /api/v1/auth/refresh` - 刷新Token
- `GET /api/v1/auth/sessions` - 获取活跃会话

### 分组管理
- `GET /api/v1/groups` - 获取分组列表（支持分页、筛选）
- `POST /api/v1/groups` - 创建分组
- `PUT /api/v1/groups/:id` - 更新分组
- `DELETE /api/v1/groups/:id` - 删除分组
- `POST /api/v1/groups/:id/regenerate-code` - 重新生成激活码
- `POST /api/v1/groups/:id/generate-subaccount-token` - 生成子账户Token
  - **功能说明**：为指定分组生成子账户登录Token，用于在新标签页自动登录子账户界面
  - **权限要求**：
    - 管理员：可以为任何分组生成Token
    - 普通用户：只能为自己管理的分组生成Token
  - **请求参数**：`id` (路径参数) - 分组ID
  - **响应示例**：
    ```json
    {
      "code": 1000,
      "message": "生成成功",
      "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
      }
    }
    ```
- `GET /api/v1/groups/categories` - 获取分组分类列表
- `POST /api/v1/groups/batch/delete` - 批量删除分组
- `POST /api/v1/groups/batch/update` - 批量更新分组

### Line账号管理
- `GET /api/v1/line-accounts` - 获取Line账号列表（支持分页、筛选）
- `POST /api/v1/line-accounts` - 创建Line账号
- `PUT /api/v1/line-accounts/:id` - 更新Line账号
- `DELETE /api/v1/line-accounts/:id` - 删除Line账号（软删除）
- `POST /api/v1/line-accounts/:id/generate-qr` - 生成二维码（Line添加好友链接）

## 🔌 WebSocket 接口文档

> **注意**：Swagger UI 主要支持 REST API，WebSocket 接口无法在 Swagger 中直接测试。WebSocket 接口的详细文档请参考独立的协议文档。

### WebSocket 连接端点

1. **Windows客户端连接**
   - 连接地址：`ws://localhost:8080/api/ws/client`
   - 认证方式：激活码 + Token（通过查询参数传递）
   - 连接示例：`ws://localhost:8080/api/ws/client?activation_code=ABC123&token=xxx`
   - 用途：Windows客户端上报数据（Line账号、进线、客户信息等）

2. **前端看板连接**
   - 连接地址：`ws://localhost:8080/api/ws/dashboard`
   - 认证方式：JWT Token（通过 HTTP Header 传递）
   - 用途：前端实时接收数据更新（账号状态、进线统计等）

### 详细协议文档

📖 **完整的 WebSocket 协议文档**：
- 🌐 **在线查看**：启动服务后访问 `http://localhost:8080/docs/websocket`
- 📄 **Markdown 文档**：项目根目录下的 [`Windows客户端交互协议.md`](../../Windows客户端交互协议.md)

该文档包含：
- ✅ 完整的消息类型定义
- ✅ 客户端 → 服务器消息格式（心跳、同步账号、进线上报等）
- ✅ 服务器 → 客户端消息格式（认证结果、同步结果、状态更新等）
- ✅ 数据归属规则和去重逻辑
- ✅ 完整的交互流程示例
- ✅ 错误处理和重连机制

### 快速参考

#### 客户端发送的消息类型

| 消息类型 | 说明 | 触发时机 |
|---------|------|---------|
| `heartbeat` | 心跳包 | 每60秒发送一次 |
| `sync_line_accounts` | 同步Line账号列表 | 连接成功后或账号变化时 |
| `incoming` | 上报进线数据 | 检测到有人加好友时 |
| `customer_sync` | 同步客户信息 | 在Line上为客户添加画像时 |
| `follow_up_sync` | 同步跟进记录 | 在Line上添加跟进记录时 |
| `account_status_change` | 账号状态变化 | Line账号登录或退出时 |

#### 服务器发送的消息类型

| 消息类型 | 说明 |
|---------|------|
| `auth_success` | 认证成功 |
| `auth_error` | 认证失败 |
| `sync_result` | 账号同步结果 |
| `incoming_received` | 进线数据接收确认 |
| `account_status_change` | 账号状态更新（推送到前端） |
| `error` | 错误消息 |

### 消息格式示例

**心跳消息**：
```json
{
  "type": "heartbeat",
  "activation_code": "ABC123",
  "timestamp": 1703123456
}
```

**同步Line账号**：
```json
{
  "type": "sync_line_accounts",
  "activation_code": "ABC123",
  "data": [
    {
      "line_id": "@line001",
      "display_name": "张三",
      "platform_type": "line",
      "online_status": "online"
    }
  ]
}
```

**上报进线**：
```json
{
  "type": "incoming",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "incoming_line_id": "U123456789",
    "timestamp": "2025-12-21 10:30:00",
    "display_name": "王五"
  }
}
```

> 💡 **提示**：更多详细的消息格式、字段说明和交互流程，请查看 [`Windows客户端交互协议.md`](../../Windows客户端交互协议.md)

## 🚀 快速开始

1. 确保后端服务正在运行（`go run cmd/server/main.go`）
2. 打开浏览器访问：`http://localhost:8080/swagger/index.html`
3. 使用登录接口获取 Token
4. 在 Swagger UI 中授权 Token
5. 开始测试 API！

## ⚠️ 常见问题

### 问题1：访问 `/swagger/index.html` 显示 "Failed to load API definition"
**解决方案**：
- 确保已经生成了 Swagger 文档：`swag init -g cmd/server/main.go -o docs`
- 检查 `backend/docs/` 目录下是否有 `swagger.json` 文件
- 重启后端服务

### 问题2：Swagger UI 中显示 401 Unauthorized
**解决方案**：
- 点击右上角 "Authorize" 按钮
- 输入正确的 JWT Token（格式：`Bearer {token}`）
- 确保 Token 未过期

### 问题3：找不到 `swag` 命令
**解决方案**：
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```
然后使用完整路径或添加到 PATH：
- Windows: `C:\Users\{username}\go\bin\swag.exe`
- Linux/Mac: `~/go/bin/swag`

