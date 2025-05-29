#!/bin/bash

# SleekChat Stop Script
# This script stops all SleekChat containers

set -e

echo "🛑 Stopping SleekChat application..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running!"
    exit 1
fi

# Stop and remove containers
echo "📦 Stopping containers..."
docker-compose down --remove-orphans

echo ""
echo "🧹 Cleaning up unused Docker resources..."
docker system prune -f --volumes

echo ""
echo "✅ SleekChat has been stopped successfully!"
echo ""
echo "🚀 To start the application again, run:"
echo "   ./start-https.sh"
echo ""
