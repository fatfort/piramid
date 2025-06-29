#!/bin/bash

# Production environment startup script
# This script loads configuration from .env file and starts the production containers

set -e

echo "Starting production containers with custom configuration..."

# Load environment variables from .env file
if [ -f ../.env ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
    echo "✅ Loaded environment variables from .env"
elif [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✅ Loaded environment variables from .env"
else
    echo "⚠️  .env file not found. Using default values."
    echo "💡 Run this from the project root or create a .env file"
fi

# Set defaults if not provided in .env
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-piramid-db-secure-2024}
export JWT_SECRET=${JWT_SECRET:-piramid-jwt-secret-key-production-2024}
export ENVIRONMENT=${ENVIRONMENT:-production}
export HOME_NET=${HOME_NET:-any}
export INTERFACE=${INTERFACE:-eth0}
export GRAFANA_PASSWORD=${GRAFANA_PASSWORD:-piramid-grafana-2024}
export ACME_EMAIL=${ACME_EMAIL:-admin@piramid.local}

echo "Configuration:"
echo "  Environment: $ENVIRONMENT"
echo "  Home Network: $HOME_NET"
echo "  Interface: $INTERFACE"
echo "  Email: $ACME_EMAIL"

# Stop any existing containers
docker compose down

# Start the containers with the environment variables
docker compose up -d postgres nats api frontend

echo "Waiting for containers to start..."
sleep 15

echo "Containers started. Creating user account..."

# Create the user using the seed program
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Creating admin user with email: $ADMIN_EMAIL"
    # Use the updated create-admin-user script
    if [ -f ../scripts/create-admin-user.sh ]; then
        cd .. && ./scripts/create-admin-user.sh
    else
        echo "Admin user creation script not found. Please run it manually."
    fi
else
    echo "⚠️  ADMIN_EMAIL and ADMIN_PASSWORD not set in .env file"
    echo "Please set these variables and run the create-admin-user.sh script manually"
fi

echo ""
echo "Production environment is ready!"
echo "You can access the application at: http://localhost:8080"
echo "API is available at: http://localhost:8001"
echo ""
echo "Login with the credentials from your .env file:"
echo "Email: ${ADMIN_EMAIL:-admin@piramid.local}"
echo "Password: [as configured in .env file]" 