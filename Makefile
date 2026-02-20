# ════════════════════════════════════════════════════════════
# Modeva Backend - Hybrid Docker Setup
# PostgreSQL & Redis in Docker, Air on Windows
# ════════════════════════════════════════════════════════════

.PHONY: dev services stop restart logs migrate migrate-create clean help

# Start everything (databases + local Air)
dev: services migrate
	@echo "✅ Services ready!"
	@echo "🚀 Starting Air locally (with instant hot-reload)..."
	@air

# Start only Docker services (PostgreSQL + Redis)
services:
	@echo "🐳 Starting Docker services..."
	@docker-compose up -d
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 5
	@echo "✅ Services running!"

# Stop all services
stop:
	@echo "🛑 Stopping services..."
	@docker-compose down

# Restart services
restart:
	@docker-compose restart

# View logs
logs:
	@docker-compose logs -f

# Run migrations
# Run migrations
migrate:
	@echo "📦 Running migrations..."
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout 30 sh -c 'until docker-compose exec -T postgres pg_isready -U postgres; do sleep 1; done' 2>/dev/null || sleep 5
	@echo "Running CMS migrations..."
	@wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database 'postgres://postgres:daiveed@localhost:5433/modeva_cms_backend?sslmode=disable' up" || echo "  ⚠️  CMS: Migration issue (may already be up-to-date)"
	@echo "Running Ecommerce migrations..."
	@wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database 'postgres://postgres:daiveed@localhost:5433/modeva_ecommerce?sslmode=disable' up" || echo "  ⚠️  Ecommerce: Migration issue (may already be up-to-date)"
	@echo "✅ Migrations complete"


# Create migration
migrate-create:
	@echo "Select database:"
	@echo "  1) CMS"
	@echo "  2) Ecommerce"
	@read -p "Enter choice [1-2]: " choice; \
	read -p "Enter migration name: " name; \
	if [ "$$choice" = "1" ]; then \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate create -ext sql -dir ./migrations/cms -seq $$name"; \
	elif [ "$$choice" = "2" ]; then \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate create -ext sql -dir ./migrations/ecommerce -seq $$name"; \
	fi

# Clean everything
clean:
	@echo "⚠️  This will remove all data!"
	@read -p "Are you sure? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		docker-compose down -v; \
		echo "✅ Cleaned"; \
	fi

# Database shell
db:
	@docker-compose exec postgres psql -U postgres

help:
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║          Modeva Backend - Hybrid Docker Setup                  ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🚀 make dev          - Start everything (instant hot-reload!)"
	@echo "🐳 make services     - Start only Docker services"
	@echo "🛑 make stop         - Stop all services"
	@echo "🔄 make restart      - Restart services"
	@echo "📋 make logs         - View Docker logs"
	@echo "📦 make migrate      - Run migrations"
	@echo "🗄️  make migrate-create - Create new migration"
	@echo "🧹 make clean        - Remove everything (including data)"
	@echo "💾 make db           - PostgreSQL shell"

.DEFAULT_GOAL := help