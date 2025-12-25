# Line账号管理系统

> **版本**: v2.0  
> **技术栈**: Go + Vue3 + Element Plus + PostgreSQL + Redis

## 📋 项目概述

Line账号分组管理与进线统计系统，支持分组管理、账号监控、进线统计、底库管理、客户管理等功能。

在cursor 协助下进行实现的
## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)

### 后端初始化

1. **进入后端目录**
```bash
cd backend
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑 .env 文件，配置数据库和Redis连接信息
```

4. **初始化数据库**
```bash
# 使用PostgreSQL客户端执行迁移脚本
psql -U lineuser -d line_management -f migrations/001_init_schema.sql
psql -U lineuser -d line_management -f migrations/002_init_admin.sql

# 或者使用Go脚本创建管理员账号
go run scripts/create_admin.go
```

5. **运行服务**
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

6. **测试服务**
```bash
# 健康检查
curl http://localhost:8080/health

# 或使用PowerShell
Invoke-WebRequest -Uri http://localhost:8080/health
```

**✅ 后端初始化已完成！**
- ✅ Go项目结构已创建
- ✅ Gin框架已配置
- ✅ GORM已集成
- ✅ PostgreSQL连接已配置
- ✅ Redis连接已配置（可选）
- ✅ 环境变量管理（viper）已配置
- ✅ 日志系统（zap）已配置
- ✅ 数据库迁移脚本已创建并执行
- ✅ 14张数据表已创建（包含分区表）
- ✅ 触发器、视图、函数已创建
- ✅ 初始管理员账号已创建（用户名: admin, 密码: admin123）

### 前端初始化

1. **进入前端目录**
```bash
cd frontend
```

2. **安装依赖**
```bash
npm install
```

3. **运行开发服务器**
```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动

### Docker部署

#### 快速部署（推荐）

**使用部署脚本（自动配置数据库密码123456，Redis无密码）：**

```bash
# Linux/macOS
./deploy.sh

# Windows
deploy.bat
```

#### 手动部署

1. **配置环境变量**
```bash
# 创建.env文件
cat > .env << EOF
# 数据库配置
POSTGRES_PASSWORD=123456

# Redis配置（无密码）
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-please
EOF
```

2. **启动服务**
```bash
# 开发环境
docker-compose up -d postgres redis backend frontend

# 生产环境
docker-compose --profile production up -d
```

#### 生产环境部署

1. **配置环境变量**
```bash
# 复制部署环境变量模板
cp env.deployment.example .env

# 编辑.env文件，设置数据库密码和其他敏感信息
# 重要：生产环境请修改默认密码！
```

2. **启动生产环境服务**
```bash
docker-compose --profile production up -d
```

3. **查看服务状态**
```bash
docker-compose ps
```

4. **查看日志**
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
```

#### 访问地址

- **开发环境**：
  - 前端：http://localhost:8081
  - 后端API：http://localhost:8080
  - Swagger文档：http://localhost:8080/swagger/index.html

- **生产环境**（需要配置NGINX_DOMAIN环境变量）：
  - 前端：https://${NGINX_DOMAIN}
  - 后端API：https://${NGINX_DOMAIN}/api/v1
  - Swagger文档：https://${NGINX_DOMAIN}/swagger/index.html
  - WebSocket文档：https://${NGINX_DOMAIN}/docs/websocket

#### 默认管理员账号

- 用户名：`admin`
- 密码：`admin123`
- ⚠️ **重要**：首次部署后请立即修改默认密码！

## 📁 项目结构

```
.
├── backend/              # Go后端
│   ├── cmd/            # 应用入口
│   ├── internal/       # 内部包
│   │   ├── config/     # 配置
│   │   ├── handlers/   # HTTP处理器
│   │   ├── middleware/ # 中间件
│   │   ├── models/     # 数据模型
│   │   ├── routes/     # 路由
│   │   ├── services/   # 业务逻辑
│   │   └── utils/      # 工具函数
│   ├── migrations/     # 数据库迁移
│   ├── pkg/            # 公共包
│   └── scripts/        # 脚本
├── frontend/           # Vue3前端
│   ├── src/
│   │   ├── api/        # API封装
│   │   ├── components/ # 组件
│   │   ├── router/     # 路由
│   │   ├── store/      # 状态管理
│   │   └── views/      # 页面
│   └── public/         # 静态资源
└── docker-compose.yml  # Docker配置
```

## 🔧 配置说明

### 后端配置 (.env)

```env
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=lineuser
DATABASE_PASSWORD=linepass
DATABASE_DBNAME=line_management

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT配置
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOUR=24
```

### 前端配置

前端使用Vite，配置在 `vite.config.js` 中。API代理已配置为 `/api`。

## 📊 数据库

### 初始化数据库

1. 创建数据库
```sql
CREATE DATABASE line_management;
```

2. 执行迁移脚本
```bash
psql -U lineuser -d line_management -f migrations/001_init_schema.sql
```

3. 创建管理员账号
```bash
psql -U lineuser -d line_management -f migrations/002_init_admin.sql
```

### 默认管理员账号

- 用户名: `admin`
- 密码: `admin123`

**⚠️ 生产环境请务必修改默认密码！**

## 🧪 开发

### 后端开发

```bash
# 运行开发服务器（带热重载）
go run cmd/server/main.go

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

### 前端开发

```bash
# 开发模式
npm run dev

# 构建生产版本
npm run build

# 代码检查
npm run lint
```

## 📝 API文档

启动服务后，访问 `http://localhost:8080/swagger/index.html` 查看API文档。

## 🐛 问题排查

### 数据库连接失败

1. 检查PostgreSQL服务是否运行
2. 检查 `.env` 中的数据库配置
3. 检查数据库用户权限

### Redis连接失败

1. 检查Redis服务是否运行
2. 检查 `.env` 中的Redis配置

### 前端无法连接后端

1. 检查后端服务是否运行
2. 检查 `vite.config.js` 中的代理配置
3. 检查CORS配置

## 📄 许可证

MIT License

## 👥 贡献

欢迎提交Issue和Pull Request！
