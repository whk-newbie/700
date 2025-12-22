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
- `GET /api/v1/groups/categories` - 获取分组分类列表
- `POST /api/v1/groups/batch/delete` - 批量删除分组
- `POST /api/v1/groups/batch/update` - 批量更新分组

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

