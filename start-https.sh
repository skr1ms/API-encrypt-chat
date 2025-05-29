#!/bin/bash

# SleekChat HTTPS Launcher
# This script starts the SleekChat application with HTTPS support

set -e

echo "🚀 Starting SleekChat with HTTPS support..."
echo ""
echo "📋 Application URLs:"
echo "  - Main App (HTTPS): https://localhost"
echo "  - Backend API:      https://localhost/api"
echo "  - WebSocket:        wss://localhost/ws"
echo ""
echo "⚠️  Security Notice:"
echo "   You may see a browser security warning because we're using"
echo "   a self-signed certificate for development. This is normal."
echo "   Click 'Advanced' → 'Proceed to localhost' to continue."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running!"
    echo "   Please start Docker Desktop and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ Error: docker-compose is not installed!"
    echo "   Please install Docker Compose and try again."
    exit 1
fi

echo "🛑 Stopping existing containers..."
docker-compose down --remove-orphans

echo ""
echo "🔧 Building and starting containers with HTTPS..."
docker-compose up --build -d

echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "📝 Checking container logs..."
echo "   Backend logs:"
docker-compose logs backend --tail=10

echo ""
echo "   Frontend logs:"
docker-compose logs frontend --tail=10

echo ""
echo "✅ SleekChat is starting up!"
echo ""
echo "🌐 Access your application:"
echo "   → https://localhost"
echo ""
echo "📊 To monitor logs in real-time, run:"
echo "   docker-compose logs -f"
echo ""
echo "🛑 To stop the application, run:"
echo "   docker-compose down"
echo ""
