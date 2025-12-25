# 域名配置指南

## 🔒 安全说明

**域名配置已改为使用环境变量，不会出现在GitHub代码中。**

所有域名配置都通过 `.env` 文件管理，该文件已在 `.gitignore` 中，不会被提交到Git。

## 📋 配置步骤

### 1. 创建环境变量文件

```bash
# 复制示例文件
cp env.deployment.example .env
```

### 2. 配置域名

编辑 `.env` 文件，设置你的域名：

```bash
# 域名配置
NGINX_DOMAIN=your-actual-domain.com
SSL_DOMAIN=your-actual-domain.com  # 可选，默认使用NGINX_DOMAIN
```

### 3. 启动服务

```bash
# 生产环境
docker-compose --profile production up -d --build
```

## 🔧 工作原理

1. **Nginx配置**：使用模板文件 `nginx/nginx.conf.template`
2. **环境变量替换**：启动时自动将 `${NGINX_DOMAIN}` 替换为实际域名
3. **SSL证书**：使用 `SSL_DOMAIN` 环境变量自动生成证书

## 📁 相关文件

- **nginx/nginx.conf.template** - Nginx配置模板（使用环境变量）
- **nginx/entrypoint.sh** - 启动脚本（自动替换环境变量）
- **.env** - 环境变量文件（不提交到Git）
- **env.deployment.example** - 环境变量示例（可提交到Git）

## ⚠️ 重要提醒

- ✅ `.env` 文件已在 `.gitignore` 中，不会被提交
- ✅ 所有硬编码域名已移除
- ✅ 使用环境变量配置，安全可靠
- ⚠️ 不要将包含真实域名的 `.env` 文件提交到Git

## 🚀 快速开始

```bash
# 1. 配置域名
echo "NGINX_DOMAIN=your-domain.com" >> .env
echo "SSL_DOMAIN=your-domain.com" >> .env

# 2. 启动服务
docker-compose --profile production up -d --build

# 3. 访问
# https://your-domain.com
```

