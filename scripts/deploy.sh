#!/bin/bash

# Line账号管理系统 - 部署脚本
# 数据库密码: 123456, Redis无密码

echo "Line账号管理系统部署脚本"
echo "数据库密码: 123456"
echo "Redis密码: 无"
echo ""

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "[错误] Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查docker-compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "[错误] docker-compose 未安装，请先安装 docker-compose"
    exit 1
fi

echo "[成功] Docker 环境检查通过"

# 创建.env文件（如果不存在）
if [ ! -f ".env" ]; then
    echo "[信息] 创建 .env 配置文件..."
    
    # 询问域名配置（生产环境需要）
    echo ""
    read -p "请输入域名（生产环境使用，留空则使用localhost）: " domain
    
    cat > .env << EOF
# 数据库配置
POSTGRES_PASSWORD=123456

# Redis配置（无密码）
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-please

# 域名配置（生产环境）
NGINX_DOMAIN=${domain:-localhost}
SSL_DOMAIN=${domain:-localhost}

# 其他配置
GIN_MODE=release
SERVER_PORT=8080
EOF
    echo "[成功] .env 文件已创建"
else
    echo "[信息] .env 文件已存在，跳过创建"
    
    # 检查是否配置了域名
    if ! grep -q "NGINX_DOMAIN" .env 2>/dev/null; then
        echo ""
        read -p "检测到.env文件但未配置域名，是否现在配置？(y/n): " config_domain
        if [ "$config_domain" = "y" ]; then
            read -p "请输入域名（留空则使用localhost）: " domain
            echo "" >> .env
            echo "# 域名配置（生产环境）" >> .env
            echo "NGINX_DOMAIN=${domain:-localhost}" >> .env
            echo "SSL_DOMAIN=${domain:-localhost}" >> .env
            echo "[成功] 域名配置已添加"
        fi
    fi
fi

echo ""
echo "[信息] 启动服务..."

# 询问用户选择环境
echo "请选择部署环境："
echo "1) 开发环境（前端直接访问）"
echo "2) 生产环境（带Nginx反向代理）"
read -p "请输入选择 (1或2): " choice

case $choice in
    1)
        echo "[信息] 启动开发环境..."
        docker-compose up -d --build postgres redis backend frontend
        echo ""
        echo "[成功] 开发环境启动完成！"
        echo "前端访问: http://localhost:8081"
        echo "后端API: http://localhost:8080"
        echo "API文档: http://localhost:8080/swagger/index.html"
        ;;
    2)
        # 检查域名配置
        if [ -f ".env" ]; then
            domain=$(grep "^NGINX_DOMAIN=" .env | cut -d'=' -f2 | tr -d '\r' | tr -d '\n')
            if [ -z "$domain" ] || [ "$domain" = "localhost" ]; then
                echo "[警告] 生产环境建议配置真实域名"
                echo "   当前域名: ${domain:-localhost}"
                echo "   如需配置，请编辑 .env 文件中的 NGINX_DOMAIN"
                echo ""
            fi
        fi
        
        echo "[信息] 启动生产环境..."
        echo "   - 自动生成SSL证书"
        echo "   - 配置Nginx反向代理"
        docker-compose --profile production up -d --build
        echo ""
        echo "[成功] 生产环境启动完成！"
        
        # 显示访问地址
        if [ -f ".env" ]; then
            domain=$(grep "^NGINX_DOMAIN=" .env | cut -d'=' -f2 | tr -d '\r' | tr -d '\n')
            if [ -n "$domain" ] && [ "$domain" != "localhost" ]; then
                echo "前端访问: https://$domain"
                echo "后端API: https://$domain/api/v1"
                echo "API文档: https://$domain/swagger/index.html"
                echo "[注意] 使用自签名证书，浏览器会显示安全警告"
            else
                echo "前端访问: http://localhost"
                echo "后端API: http://localhost:8080"
                echo "API文档: http://localhost:8080/swagger/index.html"
            fi
        else
            echo "前端访问: http://localhost"
            echo "后端API: http://localhost:8080"
        fi
        ;;
    *)
        echo "[错误] 无效选择，退出"
        exit 1
        ;;
esac

echo ""
echo "[信息] 等待服务启动..."
sleep 15

echo ""
echo "[信息] 检查服务状态..."
docker-compose ps

echo ""
echo "默认管理员账号："
echo "   用户名: admin"
echo "   密码: admin123"
echo "[重要] 请立即登录并修改默认密码！"

echo ""
echo "常用命令："
echo "   查看日志: docker-compose logs -f"
echo "   查看特定服务日志: docker-compose logs -f backend"
echo "   重启服务: docker-compose restart [service_name]"
echo "   停止服务: docker-compose down"
echo "   停止并删除数据: docker-compose down -v"
echo ""
echo "[成功] 部署完成！"
