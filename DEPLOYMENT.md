# Line账号管理系统 - 部署指南

## 📋 概述

本文档介绍如何部署Line账号分组管理与进线统计系统。

## 🐳 Docker部署（推荐）

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少2GB RAM
- 至少10GB磁盘空间

### 快速开始

1. **克隆项目**
```bash
git clone <repository-url>
cd line-management-system
```

2. **配置环境变量**
```bash
cp backend/env.example.txt backend/.env
# 编辑backend/.env文件，设置数据库密码和其他敏感信息
```

3. **启动服务**
```bash
# 开发环境（前端直接访问）
docker-compose up -d postgres redis backend frontend

# 生产环境（带Nginx反向代理）
docker-compose --profile production up -d
```

4. **访问应用**
- 前端：http://localhost
- 后端API：http://localhost:8080
- Swagger文档：http://localhost:8080/swagger/index.html
- WebSocket文档：http://localhost:8080/docs/websocket

### 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| postgres | 5432 | PostgreSQL数据库 |
| redis | 6379 | Redis缓存 |
| backend | 8080 | Go后端API服务 |
| frontend | 80 | Vue3前端应用 |
| nginx | 80/443 | Nginx反向代理（生产环境） |

## 🔧 环境配置

### 必需环境变量

#### 创建环境配置文件
```bash
# 复制部署环境变量模板
cp env.deployment.example .env

# 编辑.env文件，设置你的密码和密钥
```

#### 数据库配置
```bash
# PostgreSQL数据库密码（必需）
POSTGRES_PASSWORD=your-secure-db-password-here

# 注意：后端会自动使用与PostgreSQL容器相同的密码
# 无需手动设置DATABASE_PASSWORD环境变量
```

#### Redis配置
```bash
# Redis密码（可选，如果不需要密码可以留空）
REDIS_PASSWORD=your-redis-password-here
```

#### JWT配置
```bash
# JWT密钥（必需，使用强随机字符串）
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

### 可选环境变量

#### SSL配置（生产环境）
```bash
# 将SSL证书放置在nginx/ssl目录下
# fullchain.pem - 完整证书链
# privkey.pem - 私钥
```

## 📊 数据库初始化

系统会在首次启动时自动创建数据库表和初始数据：

1. **数据库表结构**：通过migrations目录中的SQL文件创建
2. **初始管理员账号**：
   - 用户名：admin
   - 密码：admin123
   - 角色：管理员

⚠️ **重要**：首次部署后请立即修改默认管理员密码！

## 🔒 安全配置

### 生产环境建议

1. **修改默认密码**
2. **启用SSL证书**
3. **配置防火墙**
4. **定期备份数据**
5. **监控日志**

### SSL证书配置

1. 获取Let's Encrypt证书或商业证书
2. 将证书文件放置在`nginx/ssl/`目录：
   - `fullchain.pem`：完整证书链
   - `privkey.pem`：私钥文件
3. 更新`nginx/nginx.conf`中的域名配置
4. 重启nginx服务

## 📈 监控和维护

### 日志查看

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f postgres
```

### 数据备份

```bash
# 备份PostgreSQL数据
docker exec line-postgres pg_dump -U lineuser line_management > backup.sql

# 备份Redis数据（如果需要）
docker exec line-redis redis-cli --raw KEYS "*" > redis_keys.txt
```

### 服务管理

```bash
# 重启服务
docker-compose restart backend

# 更新服务
docker-compose pull
docker-compose up -d

# 停止所有服务
docker-compose down
```

## 🚀 性能优化

### 服务器配置建议

- **CPU**：2核以上
- **内存**：4GB以上
- **磁盘**：SSD，I/O性能良好

### 并发配置

- WebSocket连接：支持800+并发
- API请求：支持1000+ QPS
- 数据库连接池：最大100连接

## 🔧 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查POSTGRES_PASSWORD环境变量
   - 确认postgres服务已启动

2. **Redis连接失败**
   - 检查REDIS_PASSWORD环境变量
   - 确认redis服务已启动

3. **前端无法访问**
   - 检查frontend服务日志
   - 确认端口80未被占用

4. **WebSocket连接失败**
   - 检查nginx配置中的WebSocket代理设置
   - 确认后端WebSocket服务正常

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost/health
curl http://localhost:8080/health
```

## 📞 支持

如遇到部署问题，请查看：
1. Docker日志：`docker-compose logs`
2. 应用日志：`backend/logs/app.log`
3. 系统文档和API文档

---

最后更新：2025-12-24
