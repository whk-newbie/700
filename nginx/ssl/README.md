# SSL证书配置说明

## 📋 概述

此目录用于存放SSL证书文件，用于Nginx的HTTPS配置。

## 📁 文件结构

```
nginx/ssl/
├── fullchain.pem    # 完整证书链（必需）
├── privkey.pem      # 私钥文件（必需）
└── README.md        # 本说明文件
```

## 🔒 获取SSL证书

### 方式一：Let's Encrypt（推荐）

1. **安装Certbot**
```bash
# Ubuntu/Debian
sudo apt install certbot

# CentOS/RHEL
sudo yum install certbot
```

2. **获取证书**
```bash
# 使用DNS验证（推荐）
sudo certbot certonly --manual --preferred-challenges dns -d yourdomain.com

# 或使用HTTP验证（需要80端口）
sudo certbot certonly --webroot -w /var/www/html -d yourdomain.com
```

3. **复制证书文件**
```bash
# 复制到项目目录
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./nginx/ssl/
```

### 方式二：商业SSL证书

1. 从证书提供商（如DigiCert、GlobalSign）购买SSL证书
2. 下载证书文件：
   - `fullchain.pem`：完整证书链（包含中间证书）
   - `privkey.pem`：私钥文件
3. 将文件放置在此目录中

### 方式三：自签名证书（开发环境）

```bash
# 生成自签名证书（仅用于开发测试）
openssl req -x509 -newkey rsa:4096 -keyout privkey.pem -out fullchain.pem -days 365 -nodes -subj "/CN=localhost"
```

## ⚙️ 配置步骤

1. **确保文件权限正确**
```bash
chmod 600 privkey.pem
chmod 644 fullchain.pem
```

2. **更新Nginx配置**
编辑`nginx/nginx.conf`文件，将`your-domain.com`替换为实际域名：
```nginx
server_name yourdomain.com;
```

3. **重启服务**
```bash
docker-compose --profile production up -d nginx
```

## 🔍 验证配置

### 检查证书有效性
```bash
# 检查证书信息
openssl x509 -in fullchain.pem -text -noout

# 检查私钥匹配性
openssl x509 -noout -modulus -in fullchain.pem | openssl md5
openssl rsa -noout -modulus -in privkey.pem | openssl md5
```

### 测试HTTPS连接
```bash
# 测试SSL连接
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# 使用curl测试
curl -I https://yourdomain.com
```

## 🔄 证书续期

### Let's Encrypt自动续期
```bash
# 设置定时任务（每月执行）
sudo crontab -e
# 添加以下行：
0 12 * * * /usr/bin/certbot renew --quiet && docker-compose --profile production restart nginx
```

### 手动续期
```bash
# 续期证书
sudo certbot renew

# 重新复制到项目目录
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./nginx/ssl/

# 重启nginx
docker-compose --profile production restart nginx
```

## ⚠️ 重要提醒

1. **私钥安全**：`privkey.pem`文件包含私钥，请妥善保管，不要提交到版本控制系统
2. **备份证书**：定期备份证书文件，以防意外丢失
3. **权限控制**：确保只有nginx进程有权限读取证书文件
4. **域名匹配**：证书的域名必须与服务器配置的域名一致

## 🐛 故障排除

### 常见错误

1. **证书不匹配域名**
   - 错误：`SSL certificate problem: certificate name mismatch`
   - 解决：确认证书域名与访问域名一致

2. **证书过期**
   - 错误：`SSL certificate expired`
   - 解决：续期或重新获取证书

3. **私钥权限错误**
   - 错误：`SSL: error:0B080074:x509 certificate routines:X509_check_private_key:key values mismatch`
   - 解决：检查私钥文件权限，确保只有所有者可读

---

最后更新：2025-12-24
