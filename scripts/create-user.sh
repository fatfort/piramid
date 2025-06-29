#!/bin/bash

# Simple script to create a regular user for PIRAMID
# Usage: ./scripts/create-user.sh <email> <password>

set -e

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: $0 <email> <password>"
    echo "Example: $0 user@fatfort.com mypassword"
    exit 1
fi

USER_EMAIL="$1"
USER_PASSWORD="$2"

# Load environment variables from .env file
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
elif [ -f "deploy/.env" ]; then
    export $(grep -v '^#' deploy/.env | xargs)
else
    echo "Error: .env file not found"
    exit 1
fi

echo "Creating user account..."
echo "Email: ${USER_EMAIL}"

# Run the create user command with custom credentials
docker run --rm \
    --network deploy_piramid-network \
    --env-file .env \
    -e ADMIN_EMAIL="${USER_EMAIL}" \
    -e ADMIN_PASSWORD="${USER_PASSWORD}" \
    -v $(pwd):/app \
    -w /app \
    golang:1.22-alpine \
    go run cmd/create_user/main.go

echo ""
echo "User created successfully!"
echo "Email: ${USER_EMAIL}"
echo "Password: ${USER_PASSWORD}"
