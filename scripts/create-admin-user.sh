#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "Loaded environment variables from .env"
else
    echo "Warning: .env file not found. Using default values."
fi

echo "Creating admin user account..."
echo "Email: ${ADMIN_EMAIL:-admin@piramid.local}"

# Create the user using the Go program in a Docker container
docker run --rm \
    --network deploy_piramid-network \
    -e DATABASE_URL="postgres://piramid:${POSTGRES_PASSWORD:-piramid-db-secure-2024}@postgres:5432/piramid?sslmode=disable" \
    -e HOME_NET="${HOME_NET:-any}" \
    -e ADMIN_EMAIL="${ADMIN_EMAIL:-admin@piramid.local}" \
    -e ADMIN_PASSWORD="${ADMIN_PASSWORD:-piramid-admin-2024}" \
    -v $(pwd):/app \
    -w /app \
    golang:1.22-alpine \
    go run cmd/create_user/main.go

echo ""
echo "Admin user created successfully!"
echo ""
echo "You can now log in at: http://localhost:8080"
echo "Email: ${ADMIN_EMAIL:-admin@piramid.local}"
echo "Password: [as configured in .env file]" 