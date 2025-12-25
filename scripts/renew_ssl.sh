#!/bin/bash

# Let's Encrypt证书自动续期脚本

DOMAIN="${SSL_DOMAIN:-${NGINX_DOMAIN:-your-domain.com}}"

echo "🔄 Let's Encrypt证书续期脚本"
echo "域名: $DOMAIN"
echo ""

# 检查certbot是否安装
if ! command -v certbot &> /dev/null; then
    echo "❌ certbot未安装，请先运行 generate_ssl.sh"
    exit 1
fi

# 续期证书
echo "🔄 正在续期证书..."
sudo certbot renew --quiet

if [ $? -eq 0 ]; then
    # 复制新证书到项目目录
    echo "📋 复制新证书文件..."
    sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem nginx/ssl/
    sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem nginx/ssl/
    sudo chmod 644 nginx/ssl/fullchain.pem
    sudo chmod 600 nginx/ssl/privkey.pem
    
    echo "✅ 证书续期成功！"
    echo "🔄 重启nginx服务..."
    docker-compose --profile production restart nginx
    echo "✅ 完成！"
else
    echo "❌ 证书续期失败"
    exit 1
fi

