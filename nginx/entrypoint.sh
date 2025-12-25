#!/bin/bash
set -e

DOMAIN="${SSL_DOMAIN:-${NGINX_DOMAIN:-localhost}}"
SSL_DIR="/etc/nginx/ssl"
NGINX_DOMAIN="${NGINX_DOMAIN:-localhost}"

# 使用envsubst替换nginx配置模板中的环境变量
if [ -f "/etc/nginx/nginx.conf.template" ]; then
    envsubst '${NGINX_DOMAIN}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf
    echo "✅ Nginx配置已生成（域名: $NGINX_DOMAIN）"
fi

# 创建SSL目录
mkdir -p "$SSL_DIR"

# 检查证书是否存在
if [ ! -f "$SSL_DIR/fullchain.pem" ] || [ ! -f "$SSL_DIR/privkey.pem" ]; then
    echo "🔐 未检测到SSL证书，正在自动生成自签名证书..."
    echo "   域名: $DOMAIN"
    echo "   ⚠️  这是自签名证书，浏览器会显示安全警告"
    
    # 生成自签名证书
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$SSL_DIR/privkey.pem" \
        -out "$SSL_DIR/fullchain.pem" \
        -subj "/C=CN/ST=State/L=City/O=Organization/CN=$DOMAIN" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        chmod 600 "$SSL_DIR/privkey.pem"
        chmod 644 "$SSL_DIR/fullchain.pem"
        echo "✅ 自签名证书生成成功！"
    else
        echo "❌ 证书生成失败，将使用HTTP模式"
        # 如果证书生成失败，可以修改nginx配置跳过SSL
    fi
else
    echo "✅ 检测到已有SSL证书，使用现有证书"
fi

# 执行nginx命令（如果传入了参数则使用参数，否则使用默认命令）
if [ $# -gt 0 ]; then
    exec "$@"
else
    exec nginx -g "daemon off;"
fi

