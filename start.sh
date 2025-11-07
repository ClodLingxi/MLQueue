#!/bin/bash

set -e

echo "======================================"
echo "  MLQueue Docker Deployment"
echo "======================================"
echo ""

# Check if docker and docker-compose are installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker compose &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo "⚠️  Please edit .env and update the JWT_SECRET and passwords!"
fi

echo "Building and starting services..."
docker compose up -d --build

echo ""
echo "Waiting for services to be healthy..."
sleep 10

# Check service health
echo ""
echo "Service Status:"
docker compose ps

echo ""
echo "======================================"
echo "  Services are running!"
echo "======================================"
echo ""
echo "Frontend:    http://localhost"
echo "Backend API: http://localhost:8080"
echo "PostgreSQL"
echo "Redis"
echo ""
echo "View logs:"
echo "  docker compose logs -f"
echo ""
echo "Stop services:"
echo "  docker compose down"
echo ""
