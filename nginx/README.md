# Nginx配置说明

## 📋 配置文件

- **nginx.conf.template** - 配置模板（使用环境变量）
- **nginx.conf** - 已废弃，保留用于参考
- **Dockerfile** - 自定义nginx镜像，支持自动生成SSL证书
- **entrypoint.sh** - 启动脚本，自动生成证书和配置

## 🔧 配置方式

### 使用环境变量配置域名

在 `.env` 文件中设置：

```bash
NGINX_DOMAIN=your-domain.com
SSL_DOMAIN=your-domain.com  # 可选，默认使用NGINX_DOMAIN
```

### 工作原理

1. **启动时**：`entrypoint.sh` 脚本会：
   - 使用 `envsubst` 将模板中的 `${NGINX_DOMAIN}` 替换为实际域名
   - 检查SSL证书是否存在，不存在则自动生成
   - 启动nginx服务

2. **SSL证书**：
   - 自动检测 `nginx/ssl/` 目录
   - 如果不存在证书，使用 `SSL_DOMAIN` 环境变量生成自签名证书
   - 证书文件：`fullchain.pem` 和 `privkey.pem`

## 🚀 使用方法

```bash
# 1. 配置域名（在.env文件中）
echo "NGINX_DOMAIN=your-domain.com" >> .env

# 2. 启动生产环境
docker-compose --profile production up -d --build

# nginx会自动：
# - 从环境变量读取域名
# - 生成nginx配置
# - 生成SSL证书（如果不存在）
```

## ⚠️ 注意事项

- 不要将包含真实域名的 `.env` 文件提交到Git
- 域名配置通过环境变量管理，不硬编码在配置文件中
- SSL证书会自动生成，无需手动操作

