# Swagger API文档自动更新说明

## 📖 Swagger文档自动生成

### ✅ 是的，API文档会在Docker部署时自动更新！

在 `backend/Dockerfile` 中已经配置了自动生成Swagger文档的步骤：

```dockerfile
# 安装swag工具用于生成Swagger文档
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 复制源代码
COPY . .

# 生成Swagger文档
RUN swag init -g cmd/server/main.go -o docs || echo "Warning: swag init failed, using existing docs"
```

## 🔄 更新流程

### Docker构建时（自动）
1. **构建阶段**：Docker构建时会自动执行 `swag init` 命令
2. **扫描代码**：swag工具会扫描所有带有Swagger注释的handler函数
3. **生成文档**：自动生成 `docs/swagger.json` 和 `docs/swagger.yaml`
4. **嵌入代码**：生成的文档会被编译到二进制文件中

### 本地开发时（手动）
如果需要更新本地Swagger文档：

```bash
cd backend
swag init -g cmd/server/main.go -o docs
```

## 📝 Swagger注释格式

在handler函数上添加Swagger注释，例如：

```go
// ProxyOpenAIAPI OpenAI API转发接口
// @Summary OpenAI API转发
// @Description 转发OpenAI API请求，前端传参格式与OpenAI文档一致，后端自动添加授权码
// @Tags 大模型调用
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schemas.OpenAIProxyRequest true "OpenAI API请求（不包含授权码）"
// @Success 200 {object} map[string]interface{} "OpenAI API响应"
// @Failure 400 {object} schemas.ErrorResponse
// @Router /llm/proxy/openai [post]
func ProxyOpenAIAPI(c *gin.Context) {
    // ...
}
```

## 🔍 查看文档

部署后访问：
- **开发环境**: `http://localhost:8080/swagger/index.html`
- **生产环境**: `https://your-domain.com/swagger/index.html`

## ⚠️ 注意事项

1. **注释必须正确**：Swagger注释必须符合swag格式，否则生成会失败
2. **失败处理**：如果swag init失败，会使用现有的docs文件（不会中断构建）
3. **文档更新**：每次Docker构建都会重新生成文档，确保与代码同步
4. **手动更新**：如果修改了Swagger注释，需要重新构建Docker镜像才能看到更新

## 🛠️ 故障排查

### 问题：Swagger文档没有更新

**检查步骤**：
1. 确认Docker构建日志中是否有 `swag init` 的执行记录
2. 检查是否有警告信息：`Warning: swag init failed`
3. 查看 `backend/docs/` 目录中的文件时间戳

**解决方法**：
```bash
# 手动重新生成文档
cd backend
swag init -g cmd/server/main.go -o docs

# 重新构建Docker镜像
docker-compose build backend
```

### 问题：Swagger UI显示404

**检查步骤**：
1. 确认 `SWAGGER_ENABLE=true` 在环境变量中
2. 检查路由是否正确配置：`/swagger/*any`
3. 查看后端日志是否有错误

## 📚 相关文件

- Swagger文档生成：`backend/Dockerfile` (第28行)
- Swagger路由配置：`backend/internal/routes/routes.go` (SetupSwagger函数)
- Swagger文档文件：`backend/docs/` 目录
- Swagger配置：`backend/cmd/server/main.go` (SwaggerInfo)

