.PHONY: help setup start stop restart logs clean build test dev create-user

# Default target
help: ## Show this help message
	@echo "PIRAMID - Network IDS Management Commands"
	@echo "========================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Setup environment and configuration files
	@echo "🔧 Setting up PIRAMID environment..."
	@if [ ! -f .env ]; then \
		echo "📄 Creating .env file from template..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env file with your configuration!"; \
		echo "⚠️  At minimum, change: POSTGRES_PASSWORD, JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD"; \
	else \
		echo "✅ .env file already exists"; \
	fi
	@chmod +x scripts/*.sh

start: ## Start all services
	@echo "🚀 Starting PIRAMID system..."
	@./scripts/start.sh

stop: ## Stop all services
	@echo "🛑 Stopping PIRAMID system..."
	@docker compose -f deploy/docker-compose.yml down

restart: stop start ## Restart all services

logs: ## View logs from all services
	@docker compose -f deploy/docker-compose.yml logs -f

logs-api: ## View API logs only
	@docker compose -f deploy/docker-compose.yml logs -f api

logs-db: ## View database logs only
	@docker compose -f deploy/docker-compose.yml logs -f postgres

status: ## Show status of all services
	@echo "📊 Service Status:"
	@docker compose -f deploy/docker-compose.yml ps
	@echo ""
	@echo "🏥 Health Check:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ API not responding"

clean: ## Stop services and remove volumes (destructive!)
	@echo "🧹 Cleaning up PIRAMID system..."
	@read -p "This will delete all data. Are you sure? (y/N) " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		docker compose -f deploy/docker-compose.yml down -v; \
		docker system prune -f; \
		echo "✅ Cleanup completed"; \
	else \
		echo ""; \
		echo "❌ Cleanup cancelled"; \
	fi

create-user: ## Create admin user account
	@echo "👤 Creating admin user..."
	@./scripts/create-admin-user.sh

build: ## Build Docker images
	@echo "🏗️  Building Docker images..."
	@docker build -f deploy/Dockerfile.api -t piramid-api .
	@docker build -f deploy/Dockerfile.eve2nats -t piramid-eve2nats .

dev: ## Start development environment (core services only)
	@echo "🚀 Starting development environment..."
	@docker compose -f deploy/docker-compose.yml up -d postgres nats
	@echo "✅ Core services started (postgres, nats)"
	@echo "💡 Run 'go run cmd/api/main.go' to start API locally"
	@echo "💡 Run 'cd web && npm run dev' to start frontend locally"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test ./...

production: ## Start production environment with all services
	@echo "🏭 Starting production environment..."
	@docker compose -f deploy/docker-compose.yml --profile production up -d

observability: ## Start with observability (Grafana)
	@echo "📊 Starting with observability stack..."
	@docker compose -f deploy/docker-compose.yml --profile observability up -d

# Database management
db-reset: ## Reset database (destructive!)
	@echo "🗄️  Resetting database..."
	@read -p "This will delete all database data. Are you sure? (y/N) " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		docker compose -f deploy/docker-compose.yml stop postgres; \
		docker compose -f deploy/docker-compose.yml rm -f postgres; \
		docker volume rm deploy_postgres_data || true; \
		docker compose -f deploy/docker-compose.yml up -d postgres; \
		echo "✅ Database reset completed"; \
	else \
		echo ""; \
		echo "❌ Database reset cancelled"; \
	fi

db-backup: ## Backup database
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@docker compose -f deploy/docker-compose.yml exec postgres pg_dump -U piramid piramid > backups/piramid_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/ directory"

# Quick commands
quick-start: setup start ## Setup and start in one command
quick-clean: clean setup ## Clean and setup in one command 