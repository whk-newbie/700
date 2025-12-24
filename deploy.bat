@echo off
REM Line账号管理系统 - Windows部署脚本
REM 数据库密码: 123456, Redis无密码

echo 🚀 Line账号管理系统部署脚本
echo 数据库密码: 123456
echo Redis密码: 无
echo.

REM 检查Docker是否安装
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker 未安装，请先安装 Docker
    pause
    exit /b 1
)

REM 检查docker-compose是否安装
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ docker-compose 未安装，请先安装 docker-compose
    pause
    exit /b 1
)

echo ✅ Docker 环境检查通过

REM 创建.env文件（如果不存在）
if not exist ".env" (
    echo 📝 创建 .env 配置文件...
    (
        echo # 数据库配置
        echo POSTGRES_PASSWORD=123456
        echo.
        echo # Redis配置（无密码）
        echo REDIS_PASSWORD=
        echo.
        echo # JWT配置
        echo JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-please
        echo.
        echo # 其他配置
        echo GIN_MODE=release
        echo SERVER_PORT=8080
    ) > .env
    echo ✅ .env 文件已创建
) else (
    echo ℹ️ .env 文件已存在，跳过创建
)

echo.
echo 🔧 启动服务...

REM 询问用户选择环境
echo 请选择部署环境：
echo 1^) 开发环境（前端直接访问）
echo 2^) 生产环境（带Nginx反向代理）
set /p choice="请输入选择 (1或2): "

if "%choice%"=="1" (
    echo 🚀 启动开发环境...
    docker-compose up -d postgres redis backend frontend
    goto :success
) else if "%choice%"=="2" (
    echo 🚀 启动生产环境...
    docker-compose --profile production up -d
    goto :success
) else (
    echo ❌ 无效选择，退出
    pause
    exit /b 1
)

:success
echo.
echo ✅ 服务启动完成！
echo ⏳ 等待服务启动...
timeout /t 10 /nobreak >nul

echo.
echo 🔍 检查服务状态...
docker-compose ps

echo.
echo 📋 默认管理员账号：
echo    用户名: admin
echo    密码: admin123
echo ⚠️  重要: 请立即登录并修改默认密码！

echo.
echo 📖 查看日志: docker-compose logs -f
echo 🛑 停止服务: docker-compose down
echo.
echo 🎉 部署完成！
pause
