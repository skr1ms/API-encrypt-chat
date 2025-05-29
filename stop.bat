@echo off
echo ========================================
echo       SleekChat - Stop Containers
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

echo Stopping SleekChat containers...
docker-compose down

echo.
echo Removing unused Docker resources...
docker system prune -f --volumes

echo.
echo Checking remaining containers...
docker-compose ps

echo.
echo ========================================
echo    SleekChat containers stopped!
echo ========================================
echo.
echo All containers have been stopped and cleaned up.
echo To start the application again, run:
echo   start-https.bat (for HTTPS version)
echo.
pause
