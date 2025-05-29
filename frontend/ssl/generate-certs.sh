#!/bin/bash

# Скрипт для генерации SSL сертификатов
# Примечание: Этот скрипт НЕ используется в Docker сборке
# Сертификаты генерируются автоматически в Dockerfile
# Данный скрипт предоставлен для справки и ручной генерации

echo "========================================="
echo "  SleekChat SSL Certificate Generator"
echo "========================================="
echo

# Создаем директорию для сертификатов
mkdir -p /etc/nginx/ssl

echo "Generating SSL certificates for localhost..."

# Генерируем приватный ключ RSA-2048
openssl genrsa -out /etc/nginx/ssl/localhost.key 2048

# Генерируем самоподписанный сертификат
openssl req -new -x509 -key /etc/nginx/ssl/localhost.key -out /etc/nginx/ssl/localhost.crt -days 365 \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=SleekChat/OU=Development/CN=localhost"

# Устанавливаем правильные права доступа
chmod 600 /etc/nginx/ssl/localhost.key
chmod 644 /etc/nginx/ssl/localhost.crt

echo
echo "SSL certificates generated successfully!"
echo "- Private key: /etc/nginx/ssl/localhost.key"
echo "- Certificate: /etc/nginx/ssl/localhost.crt"
echo "- Validity: 365 days"
echo "- Algorithm: RSA-2048"
echo "- Domain: localhost"
echo
echo "Note: In Docker, certificates are generated automatically"
echo "during the build process. This script is for reference only."
