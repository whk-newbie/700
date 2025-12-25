#!/bin/bash

# 快速生成自签名SSL证书（最简单的方式）

DOMAIN="${SSL_DOMAIN:-${NGINX_DOMAIN:-your-domain.com}}"

echo "🔐 快速生成自签名SSL证书"
echo "域名: $DOMAIN"
echo ""

# 创建SSL目录
mkdir -p nginx/ssl

# 生成自签名证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout nginx/ssl/privkey.pem \
    -out nginx/ssl/fullchain.pem \
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=$DOMAIN" 2>/dev/null

if [ $? -eq 0 ]; then
    chmod 600 nginx/ssl/privkey.pem
    chmod 644 nginx/ssl/fullchain.pem
    echo "✅ 自签名证书生成成功！"
    echo ""
    echo "📋 证书文件："
    echo "   - nginx/ssl/fullchain.pem"
    echo "   - nginx/ssl/privkey.pem"
    echo ""
    echo "⚠️  注意：浏览器会显示安全警告，这是正常的"
    echo "   点击'高级' -> '继续访问'即可"
    echo ""
    echo "🚀 现在可以启动生产环境："
    echo "   docker-compose --profile production up -d"
else
    echo "❌ 证书生成失败，请检查是否安装了OpenSSL"
    echo "   Ubuntu/Debian: sudo apt-get install openssl"
    echo "   CentOS/RHEL: sudo yum install openssl"
    exit 1
fi

