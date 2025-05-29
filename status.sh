#!/bin/bash

# SleekChat Status Script
# This script shows the status of SleekChat containers

set -e

echo "📊 SleekChat Application Status"
echo "==============================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running!"
    exit 1
fi

echo "🐳 Docker Containers:"
docker-compose ps

echo ""
echo "📝 Recent Logs (Backend):"
echo "-------------------------"
docker-compose logs backend --tail=15

echo ""
echo "📝 Recent Logs (Frontend):"
echo "--------------------------"
docker-compose logs frontend --tail=10

echo ""
echo "📝 Recent Logs (Database):"
echo "--------------------------"
docker-compose logs postgres --tail=5

echo ""
echo "🌐 Application Access:"
echo "  → Main App: https://localhost"
echo "  → Backend:  https://localhost/api"
echo ""
echo "🔧 Management Commands:"
echo "  → Start:    ./start-https.sh"
echo "  → Stop:     ./stop.sh"
echo "  → Logs:     docker-compose logs -f"
echo ""
