#!/bin/bash

# SSL证书自动生成脚本
# 支持Let's Encrypt和自签名证书

DOMAIN="${SSL_DOMAIN:-${NGINX_DOMAIN:-your-domain.com}}"
EMAIL="${SSL_EMAIL:-admin@example.com}"  # 修改为你的邮箱

echo "🔐 SSL证书生成脚本"
echo "域名: $DOMAIN"
echo ""

# 创建SSL目录
mkdir -p nginx/ssl

# 检查是否已有证书
if [ -f "nginx/ssl/fullchain.pem" ] && [ -f "nginx/ssl/privkey.pem" ]; then
    echo "⚠️  检测到已有SSL证书"
    read -p "是否重新生成？(y/n): " regenerate
    if [ "$regenerate" != "y" ]; then
        echo "✅ 使用现有证书"
        exit 0
    fi
fi

echo "请选择证书类型："
echo "1) Let's Encrypt (免费，需要域名已解析)"
echo "2) 自签名证书 (仅用于测试)"
read -p "请输入选择 (1/2): " cert_type

case $cert_type in
    1)
        echo ""
        echo "📋 使用Let's Encrypt获取证书"
        echo "⚠️  需要确保："
        echo "   1. 域名 $DOMAIN 已解析到服务器IP"
        echo "   2. 服务器80和443端口已开放"
        echo ""
        read -p "确认继续？(y/n): " confirm
        if [ "$confirm" != "y" ]; then
            echo "❌ 已取消"
            exit 1
        fi

        # 检查certbot是否安装
        if ! command -v certbot &> /dev/null; then
            echo "📦 安装certbot..."
            if command -v apt-get &> /dev/null; then
                sudo apt-get update
                sudo apt-get install -y certbot
            elif command -v yum &> /dev/null; then
                sudo yum install -y certbot
            else
                echo "❌ 无法自动安装certbot，请手动安装"
                exit 1
            fi
        fi

        # 使用standalone模式获取证书（需要先停止nginx）
        echo "🛑 请先停止nginx服务（如果正在运行）"
        echo "   命令: docker-compose --profile production stop nginx"
        read -p "已停止nginx？(y/n): " nginx_stopped
        if [ "$nginx_stopped" != "y" ]; then
            echo "❌ 请先停止nginx"
            exit 1
        fi

        # 获取证书
        echo "🔐 正在获取Let's Encrypt证书..."
        sudo certbot certonly --standalone \
            -d "$DOMAIN" \
            --email "$EMAIL" \
            --agree-tos \
            --non-interactive \
            --preferred-challenges http

        if [ $? -eq 0 ]; then
            # 复制证书到项目目录
            echo "📋 复制证书文件..."
            sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem nginx/ssl/
            sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem nginx/ssl/
            sudo chmod 644 nginx/ssl/fullchain.pem
            sudo chmod 600 nginx/ssl/privkey.pem
            echo "✅ Let's Encrypt证书获取成功！"
        else
            echo "❌ 证书获取失败，使用自签名证书作为备用"
            cert_type=2
        fi
        ;;
    2)
        echo ""
        echo "🔐 生成自签名证书（仅用于测试）"
        echo "⚠️  浏览器会显示安全警告，这是正常的"
        echo ""
        
        # 生成自签名证书
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout nginx/ssl/privkey.pem \
            -out nginx/ssl/fullchain.pem \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=$DOMAIN"
        
        chmod 600 nginx/ssl/privkey.pem
        chmod 644 nginx/ssl/fullchain.pem
        
        echo "✅ 自签名证书生成成功！"
        echo "⚠️  注意：浏览器会显示安全警告，点击'高级'->'继续访问'即可"
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac

echo ""
echo "📋 证书文件位置："
echo "   - nginx/ssl/fullchain.pem"
echo "   - nginx/ssl/privkey.pem"
echo ""
echo "✅ SSL证书配置完成！"
echo ""
echo "🚀 现在可以启动生产环境："
echo "   docker-compose --profile production up -d"

