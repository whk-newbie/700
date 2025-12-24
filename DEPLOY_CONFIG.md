# 部署配置说明

## 数据库密码设置

### 当前配置
- **数据库密码**: `123456`
- **Redis密码**: 无密码（留空）

### 配置文件

创建 `.env` 文件并设置以下内容：

```bash
# 数据库配置
POSTGRES_PASSWORD=123456

# Redis配置（无密码）
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-please
```

### 启动服务

```bash
# 开发环境
docker-compose up -d

# 生产环境
docker-compose --profile production up -d
```

### 验证配置

启动后，可以通过以下方式验证：

```bash
# 检查服务状态
docker-compose ps

# 查看日志
docker-compose logs backend

# 测试数据库连接
docker-compose exec backend /app/main --help
```

### 默认管理员账号

- **用户名**: `admin`
- **密码**: `admin123`
- ⚠️ **重要**: 首次登录后请立即修改默认密码！

### 访问地址

- **前端**: http://localhost
- **API文档**: http://localhost:8080/swagger/index.html
- **WebSocket文档**: http://localhost:8080/docs/websocket
