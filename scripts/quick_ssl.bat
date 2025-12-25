@echo off
REM 快速生成自签名SSL证书 (Windows)

set DOMAIN=%SSL_DOMAIN%
if "%DOMAIN%"=="" set DOMAIN=%NGINX_DOMAIN%
if "%DOMAIN%"=="" set DOMAIN=your-domain.com

echo 🔐 快速生成自签名SSL证书
echo 域名: %DOMAIN%
echo.

REM 创建SSL目录
if not exist "nginx\ssl" mkdir nginx\ssl

REM 检查OpenSSL
where openssl >nul 2>&1
if errorlevel 1 (
    echo ❌ 未找到OpenSSL
    echo.
    echo 请安装OpenSSL：
    echo 1. 下载: https://slproweb.com/products/Win32OpenSSL.html
    echo 2. 安装后添加到PATH环境变量
    echo 3. 或使用Git Bash（已包含OpenSSL）
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
echo.
echo 📋 证书文件：
echo    - nginx\ssl\fullchain.pem
echo    - nginx\ssl\privkey.pem
echo.
echo ⚠️  注意：浏览器会显示安全警告，这是正常的
echo    点击'高级' -^> '继续访问'即可
echo.
echo 🚀 现在可以启动生产环境：
echo    docker-compose --profile production up -d
pause

