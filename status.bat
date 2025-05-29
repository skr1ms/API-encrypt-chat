@echo off
echo ========================================
echo       SleekChat - Container Status
echo ========================================
echo.

echo Checking Docker status...
docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not installed or not running!
    echo Please install Docker Desktop and make sure it's running.
    pause
    exit /b 1
)

echo Docker is running
echo.

echo ========================================
echo           Container Status
echo ========================================
docker-compose ps

echo.
echo ========================================
echo           Container Health
echo ========================================
echo.

echo Frontend Status:
docker-compose exec -T frontend curl -f http://localhost:80/health 2>nul
if errorlevel 1 (
    echo Frontend: NOT HEALTHY
) else (
    echo Frontend: HEALTHY
)

echo.
echo Backend Status:
docker-compose exec -T backend curl -f http://localhost:8080/health 2>nul
if errorlevel 1 (
    echo Backend: NOT HEALTHY
) else (
    echo Backend: HEALTHY
)

echo.
echo Database Status:
docker-compose exec -T db pg_isready -U sleek_chat_user -d sleek_chat_db 2>nul
if errorlevel 1 (
    echo Database: NOT HEALTHY
) else (
    echo Database: HEALTHY
)

echo.
echo ========================================
echo           Container Logs
echo ========================================
echo Recent logs (last 10 lines per service):
echo.

echo === Frontend Logs ===
docker-compose logs --tail=10 frontend

echo.
echo === Backend Logs ===
docker-compose logs --tail=10 backend

echo.
echo === Database Logs ===
docker-compose logs --tail=10 db

echo.
echo ========================================
echo           Resource Usage
echo ========================================
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

echo.
echo ========================================
echo Application URLs:
echo - Frontend (HTTPS): https://localhost
echo - Frontend (HTTP):  http://localhost (redirects to HTTPS)
echo - Backend API:      https://localhost/api
echo - WebSocket:        wss://localhost/ws
echo ========================================
echo.
pause
