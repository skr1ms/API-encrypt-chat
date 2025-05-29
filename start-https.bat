@echo off
echo ==========================================
echo    SleekChat - Starting with HTTPS
echo ==========================================
echo.
echo The application will be available at:
echo - HTTPS: https://localhost (main)
echo - HTTP:  http://localhost (redirects to HTTPS)
echo - Backend API: https://localhost/api
echo.
echo Note: You may see a security warning because we're using 
echo a self-signed certificate. This is normal for development.
echo Click "Advanced" and "Proceed to localhost" to continue.
echo.

echo Stopping any running containers...
docker-compose down

echo.
echo Building and starting containers with HTTPS...
docker-compose up --build -d

echo.
echo Waiting for services to start...
timeout /t 10 /nobreak >nul

echo.
echo Checking container status...
docker-compose ps

echo.
echo ==========================================
echo SleekChat is starting up!
echo ==========================================
echo.
echo Please wait a moment for all services to initialize.
echo Then open your browser and go to: https://localhost
echo.

pause
