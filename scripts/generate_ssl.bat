@echo off
REM SSL证书自动生成脚本 (Windows)
REM 支持自签名证书

set DOMAIN=%SSL_DOMAIN%
if "%DOMAIN%"=="" set DOMAIN=%NGINX_DOMAIN%
if "%DOMAIN%"=="" set DOMAIN=your-domain.com

echo 🔐 SSL证书生成脚本
echo 域名: %DOMAIN%
echo.

REM 创建SSL目录
if not exist "nginx\ssl" mkdir nginx\ssl

REM 检查是否已有证书
if exist "nginx\ssl\fullchain.pem" if exist "nginx\ssl\privkey.pem" (
    echo ⚠️  检测到已有SSL证书
    set /p regenerate="是否重新生成？(y/n): "
    if not "%regenerate%"=="y" (
        echo ✅ 使用现有证书
        exit /b 0
    )
)

echo.
echo 🔐 生成自签名证书（仅用于测试）
echo ⚠️  浏览器会显示安全警告，这是正常的
echo.

REM 检查OpenSSL是否安装
where openssl >nul 2>&1
if errorlevel 1 (
    echo ❌ 未找到OpenSSL，请先安装OpenSSL
    echo 下载地址: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

REM 生成自签名证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 ^
    -keyout nginx\ssl\privkey.pem ^
    -out nginx\ssl\fullchain.pem ^
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=%DOMAIN%"

if errorlevel 1 (
    echo ❌ 证书生成失败
    pause
    exit /b 1
)

echo.
echo ✅ 自签名证书生成成功！
echo ⚠️  注意：浏览器会显示安全警告，点击'高级'->'继续访问'即可
echo.
echo 📋 证书文件位置：
echo    - nginx\ssl\fullchain.pem
echo    - nginx\ssl\privkey.pem
echo.
echo ✅ SSL证书配置完成！
echo.
echo 🚀 现在可以启动生产环境：
echo    docker-compose --profile production up -d
pause

