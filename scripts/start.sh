#!/bin/bash

# PIRAMID Startup Script
# This script starts the PIRAMID system using environment variables from .env

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🏗️  Starting PIRAMID System${NC}"
echo "================================"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found. Creating from template...${NC}"
    cp .env.example .env
    echo -e "${RED}❗ Please edit .env file with your configuration before running again!${NC}"
    echo -e "${RED}❗ At minimum, change POSTGRES_PASSWORD, JWT_SECRET, and admin credentials${NC}"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)
echo -e "${GREEN}✅ Loaded environment variables from .env${NC}"

# Validate required variables
if [ -z "$POSTGRES_PASSWORD" ] || [ -z "$JWT_SECRET" ] || [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo -e "${RED}❌ Required environment variables are missing!${NC}"
    echo "Please ensure .env contains: POSTGRES_PASSWORD, JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD"
    exit 1
fi

echo -e "${BLUE}🐳 Starting Docker containers...${NC}"

# Start the containers
docker compose -f deploy/docker-compose.yml up -d

echo -e "${YELLOW}⏳ Waiting for services to start...${NC}"
sleep 15

# Check if services are healthy
echo -e "${BLUE}🔍 Checking service health...${NC}"
docker compose -f deploy/docker-compose.yml ps

# Test API health
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ API is healthy${NC}"
else
    echo -e "${RED}❌ API is not responding${NC}"
    echo "Check logs with: docker compose -f deploy/docker-compose.yml logs api"
    exit 1
fi

# Create admin user
echo -e "${BLUE}👤 Creating admin user...${NC}"
./scripts/create-admin-user.sh

echo ""
echo -e "${GREEN}🎉 PIRAMID is ready!${NC}"
echo "================================"
echo -e "🌐 Web Interface: ${BLUE}http://localhost:8080${NC}"
echo -e "🔧 API Endpoint:  ${BLUE}http://localhost:8001${NC}"
echo -e "📊 Grafana:       ${BLUE}http://localhost:3000${NC} (if enabled)"
echo ""
echo -e "📧 Admin Email:    ${YELLOW}$ADMIN_EMAIL${NC}"
echo -e "🔑 Admin Password: ${YELLOW}[as configured in .env]${NC}"
echo ""
echo -e "${BLUE}📋 Useful Commands:${NC}"
echo "  View logs:    docker compose -f deploy/docker-compose.yml logs -f"
echo "  Stop system:  docker compose -f deploy/docker-compose.yml down"
echo "  Reset system: docker compose -f deploy/docker-compose.yml down -v" 