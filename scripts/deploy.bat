@echo off
chcp 65001 >nul
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
    echo.
    echo 📝 创建 .env 配置文件...
    set /p domain="请输入域名（生产环境使用，留空则使用localhost）: "
    
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
        echo # 域名配置（生产环境）
        echo NGINX_DOMAIN=%domain%
        if "%domain%"=="" echo NGINX_DOMAIN=localhost
        echo SSL_DOMAIN=%domain%
        if "%domain%"=="" echo SSL_DOMAIN=localhost
        echo.
        echo # 其他配置
        echo GIN_MODE=release
        echo SERVER_PORT=8080
    ) > .env
    echo ✅ .env 文件已创建
) else (
    echo ℹ️ .env 文件已存在，跳过创建
    
    REM 检查是否配置了域名
    findstr /C:"NGINX_DOMAIN" .env >nul 2>&1
    if errorlevel 1 (
        echo.
        set /p config_domain="检测到.env文件但未配置域名，是否现在配置？(y/n): "
        if /i "%config_domain%"=="y" (
            set /p domain="请输入域名（留空则使用localhost）: "
            echo. >> .env
            echo # 域名配置（生产环境） >> .env
            echo NGINX_DOMAIN=%domain% >> .env
            if "%domain%"=="" echo NGINX_DOMAIN=localhost >> .env
            echo SSL_DOMAIN=%domain% >> .env
            if "%domain%"=="" echo SSL_DOMAIN=localhost >> .env
            echo ✅ 域名配置已添加
        )
    )
)

echo.
echo 🔧 启动服务...

REM 询问用户选择环境
echo 请选择部署环境：
echo 1^) 开发环境（前端直接访问）
echo 2^) 生产环境（带Nginx反向代理和SSL）
set /p choice="请输入选择 (1或2): "

if "%choice%"=="1" (
    echo 🚀 启动开发环境...
    docker-compose up -d --build postgres redis backend frontend
    goto :success_dev
) else if "%choice%"=="2" (
    REM 检查域名配置
    if exist ".env" (
        for /f "tokens=2 delims==" %%a in ('findstr "^NGINX_DOMAIN=" .env') do set domain=%%a
        if "%domain%"=="" set domain=localhost
        if "%domain%"=="localhost" (
            echo.
            echo ⚠️  警告: 生产环境建议配置真实域名
            echo    当前域名: %domain%
            echo    如需配置，请编辑 .env 文件中的 NGINX_DOMAIN
            echo.
        )
    )
    
    echo 🚀 启动生产环境...
    echo    - 自动生成SSL证书
    echo    - 配置Nginx反向代理
    docker-compose --profile production up -d --build
    goto :success_prod
) else (
    echo ❌ 无效选择，退出
    pause
    exit /b 1
)

:success_dev
echo.
echo ✅ 开发环境启动完成！
echo 📱 前端访问: http://localhost:8081
echo 🔗 后端API: http://localhost:8080
echo 📖 API文档: http://localhost:8080/swagger/index.html
goto :end

:success_prod
echo.
echo ✅ 生产环境启动完成！

REM 显示访问地址
if exist ".env" (
    for /f "tokens=2 delims==" %%a in ('findstr "^NGINX_DOMAIN=" .env') do set domain=%%a
    if "%domain%"=="" set domain=localhost
    if not "%domain%"=="localhost" (
        echo 📱 前端访问: https://%domain%
        echo 🔗 后端API: https://%domain%/api/v1
        echo 📖 API文档: https://%domain%/swagger/index.html
        echo ⚠️  注意: 使用自签名证书，浏览器会显示安全警告
    ) else (
        echo 📱 前端访问: http://localhost
        echo 🔗 后端API: http://localhost:8080
        echo 📖 API文档: http://localhost:8080/swagger/index.html
    )
) else (
    echo 📱 前端访问: http://localhost
    echo 🔗 后端API: http://localhost:8080
)

:end
echo.
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
