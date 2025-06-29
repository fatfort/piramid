.PHONY: help dev prod build test clean install seed logs stop restart

# Default target
help:
	@echo "PIRAMID - Predictive Intrusion Detection System"
	@echo ""
	@echo "Available commands:"
	@echo "  make dev        - Start development environment"
	@echo "  make prod       - Start production environment"
	@echo "  make build      - Build all Docker images"
	@echo "  make test       - Run all tests"
	@echo "  make clean      - Clean up containers and volumes"
	@echo "  make install    - Install dependencies"
	@echo "  make seed       - Seed the database with initial data"
	@echo "  make logs       - Show logs from all services"
	@echo "  make stop       - Stop all services"
	@echo "  make restart    - Restart all services"

# Development environment
dev:
	@echo "Starting PIRAMID development environment..."
	@cp .env.example .env 2>/dev/null || true
	docker compose -f deploy/docker-compose.yml up --build -d postgres nats api frontend
	@echo "Waiting for services to be ready..."
	@sleep 10
	@make seed
	@echo ""
	@echo "✅ PIRAMID is ready!"
	@echo "🌐 Dashboard: http://localhost:8001"
	@echo "📊 API: http://localhost:8001/health"
	@echo "📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo ""
	@echo "Demo credentials:"
	@echo "  Admin: admin@fatfort.local / admin123"
	@echo "  User:  user@fatfort.local / user123"

# Production environment with monitoring
prod:
	@echo "Starting PIRAMID production environment..."
	@cp .env.example .env 2>/dev/null || true
	docker compose -f deploy/docker-compose.yml --profile production --profile observability up --build -d
	@echo "Waiting for services to be ready..."
	@sleep 15
	@make seed
	@echo ""
	@echo "✅ PIRAMID production environment is ready!"
	@echo "🌐 Dashboard: http://localhost:8001"
	@echo "📊 Grafana: http://localhost:3000"
	@echo "🔍 Traefik: http://localhost:8080"

# Build all images
build:
	@echo "Building PIRAMID Docker images..."
	docker compose -f deploy/docker-compose.yml build

# Run tests
test:
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd web && npm test 2>/dev/null || echo "Frontend tests not configured"

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd web && npm install

# Seed database
seed:
	@echo "Seeding database..."
	docker compose -f deploy/docker-compose.yml exec -T api ./seed || \
	docker run --rm --network piramid_piramid-network \
		-e DATABASE_URL=postgres://piramid:piramid123@postgres:5432/piramid?sslmode=disable \
		piramid-api ./seed

# Show logs
logs:
	docker compose -f deploy/docker-compose.yml logs -f

# Stop all services
stop:
	docker compose -f deploy/docker-compose.yml --profile production --profile observability down

# Restart services
restart: stop
	@make dev

# Clean up everything
clean:
	@echo "Cleaning up PIRAMID environment..."
	docker compose -f deploy/docker-compose.yml --profile production --profile observability down -v
	docker system prune -f
	@echo "Cleanup complete!"

# Database migrations
migrate:
	@echo "Running database migrations..."
	docker compose -f deploy/docker-compose.yml exec api ./seed

# Backup database
backup:
	@echo "Creating database backup..."
	@mkdir -p backups
	docker compose -f deploy/docker-compose.yml exec postgres pg_dump -U piramid piramid > backups/piramid-$(shell date +%Y%m%d-%H%M%S).sql
	@echo "Backup created in backups/ directory"

# Restore database
restore:
	@echo "Usage: make restore BACKUP_FILE=backups/piramid-YYYYMMDD-HHMMSS.sql"
	@test -n "$(BACKUP_FILE)" || (echo "Please specify BACKUP_FILE=<path_to_backup>" && exit 1)
	docker compose -f deploy/docker-compose.yml exec -T postgres psql -U piramid -d piramid < $(BACKUP_FILE)

# Security scan
security:
	@echo "Running security scans..."
	@command -v trivy >/dev/null 2>&1 || (echo "Please install trivy for security scanning" && exit 1)
	trivy image piramid-api
	trivy image piramid-frontend

# Performance test
perf:
	@echo "Running performance tests..."
	@command -v ab >/dev/null 2>&1 || (echo "Please install apache2-utils for performance testing" && exit 1)
	ab -n 1000 -c 10 http://localhost:67546/health

# Update GeoIP database
update-geoip:
	@echo "Updating GeoIP database..."
	@test -n "$(GEOIP_LICENSE_KEY)" || (echo "Please set GEOIP_LICENSE_KEY environment variable" && exit 1)
	curl -o /tmp/GeoLite2-City.tar.gz "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$(GEOIP_LICENSE_KEY)&suffix=tar.gz"
	docker volume create piramid_geoip_data
	docker run --rm -v piramid_geoip_data:/data -v /tmp:/tmp alpine tar -xzf /tmp/GeoLite2-City.tar.gz -C /data --strip-components=1 