@echo off
echo Rebuilding EncryptChat application...
docker-compose down
docker-compose up --build
