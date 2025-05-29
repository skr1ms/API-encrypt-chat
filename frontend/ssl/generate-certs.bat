@echo off
REM Скрипт для генерации SSL сертификатов на Windows
REM Примечание: Этот скрипт НЕ используется в Docker сборке
REM Сертификаты генерируются автоматически в Dockerfile
REM Данный скрипт предоставлен для справки и ручной генерации

echo =========================================
echo   SleekChat SSL Certificate Generator
echo =========================================
echo.

REM Проверяем наличие OpenSSL
where openssl >nul 2>&1
if errorlevel 1 (
    echo ERROR: OpenSSL is not installed or not in PATH!
    echo Please install OpenSSL or use the Docker version.
    echo.
    echo You can download OpenSSL from:
    echo https://slproweb.com/products/Win32OpenSSL.html
    echo.
    pause
    exit /b 1
)

echo OpenSSL found. Generating SSL certificates for localhost...
echo.

REM Создаем директорию для сертификатов
if not exist "ssl" mkdir ssl

REM Генерируем приватный ключ RSA-2048
echo Generating private key...
openssl genrsa -out ssl\localhost.key 2048

REM Генерируем самоподписанный сертификат
echo Generating self-signed certificate...
openssl req -new -x509 -key ssl\localhost.key -out ssl\localhost.crt -days 365 -subj "/C=RU/ST=Moscow/L=Moscow/O=SleekChat/OU=Development/CN=localhost"

echo.
echo SSL certificates generated successfully!
echo - Private key: ssl\localhost.key
echo - Certificate: ssl\localhost.crt
echo - Validity: 365 days
echo - Algorithm: RSA-2048
echo - Domain: localhost
echo.
echo Note: In Docker, certificates are generated automatically
echo during the build process. This script is for reference only.
echo.
pause
