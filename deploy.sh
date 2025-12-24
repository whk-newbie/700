#!/bin/bash

# Line账号管理系统 - 部署脚本
# 数据库密码: 123456, Redis无密码

echo "🚀 Line账号管理系统部署脚本"
echo "数据库密码: 123456"
echo "Redis密码: 无"
echo ""

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查docker-compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose 未安装，请先安装 docker-compose"
    exit 1
fi

echo "✅ Docker 环境检查通过"

# 创建.env文件（如果不存在）
if [ ! -f ".env" ]; then
    echo "📝 创建 .env 配置文件..."
    cat > .env << EOF
# 数据库配置
POSTGRES_PASSWORD=123456

# Redis配置（无密码）
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-please

# 其他配置
GIN_MODE=release
SERVER_PORT=8080
EOF
    echo "✅ .env 文件已创建"
else
    echo "ℹ️ .env 文件已存在，跳过创建"
fi

echo ""
echo "🔧 启动服务..."

# 询问用户选择环境
echo "请选择部署环境："
echo "1) 开发环境（前端直接访问）"
echo "2) 生产环境（带Nginx反向代理）"
read -p "请输入选择 (1或2): " choice

case $choice in
    1)
        echo "🚀 启动开发环境..."
        docker-compose up -d postgres redis backend frontend
        echo ""
        echo "✅ 开发环境启动完成！"
        echo "📱 前端访问: http://localhost"
        echo "🔗 API文档: http://localhost:8080/swagger/index.html"
        ;;
    2)
        echo "🚀 启动生产环境..."
        docker-compose --profile production up -d
        echo ""
        echo "✅ 生产环境启动完成！"
        echo "📱 前端访问: http://localhost"
        echo "🔗 API文档: http://localhost:8080/swagger/index.html"
        ;;
    *)
        echo "❌ 无效选择，退出"
        exit 1
        ;;
esac

echo ""
echo "⏳ 等待服务启动..."
sleep 10

echo ""
echo "🔍 检查服务状态..."
docker-compose ps

echo ""
echo "📋 默认管理员账号："
echo "   用户名: admin"
echo "   密码: admin123"
echo "⚠️  重要: 请立即登录并修改默认密码！"

echo ""
echo "📖 查看日志: docker-compose logs -f"
echo "🛑 停止服务: docker-compose down"
echo ""
echo "🎉 部署完成！"
