# ════════════════════════════════════════════════════════════
# Modeva Backend - Hybrid Docker Setup
# PostgreSQL & Redis in Docker, Air on Windows
# ════════════════════════════════════════════════════════════

.PHONY: dev services stop restart logs migrate migrate-create migrate-neon migrate-railway clean help export-to-neon export-data-only clean-neon export-to-railway export-data-only-railway

# Load environment variables from .env (includes Neon credentials)
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Railway URLs are loaded from .env (never hardcode these - public repo!)

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

# Run migrations (local)
migrate:
	@echo "📦 Running migrations..."
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout 30 sh -c 'until docker-compose exec -T postgres pg_isready -U postgres; do sleep 1; done' 2>/dev/null || sleep 5
	@echo "Running CMS migrations..."
	@wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database '$(CMS_DB_URL)' up" || echo "  ⚠️  CMS: Migration issue (may already be up-to-date)"
	@echo "Running Ecommerce migrations..."
	@wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database '$(ECOMMERCE_DB_URL)' up" || echo "  ⚠️  Ecommerce: Migration issue (may already be up-to-date)"
	@echo "✅ Migrations complete"

# Run migrations on Neon (production databases)
migrate-neon:
	@echo "📦 Running migrations on Neon..."
	@echo "⚠️  This will run migrations on your NEON databases!"
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Running CMS migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database '$(NEON_CMS_DB_URL)' up" || echo "  ⚠️  CMS: Migration issue (may already be up-to-date)"; \
		echo "Running Ecommerce migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database '$(NEON_ECOMMERCE_DB_URL)' up" || echo "  ⚠️  Ecommerce: Migration issue (may already be up-to-date)"; \
		echo "✅ Neon migrations complete"; \
	else \
		echo "❌ Migration cancelled"; \
	fi

# Run migrations on Railway (production databases)
migrate-railway:
	@echo "📦 Running migrations on Railway..."
	@echo "⚠️  This will run migrations on your RAILWAY databases!"
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Running CMS migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database '$(RAILWAY_CMS_DB_URL)' up" || echo "  ⚠️  CMS: Migration issue (may already be up-to-date)"; \
		echo "Running Ecommerce migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database '$(RAILWAY_ECOMMERCE_DB_URL)' up" || echo "  ⚠️  Ecommerce: Migration issue (may already be up-to-date)"; \
		echo "✅ Railway migrations complete"; \
	else \
		echo "❌ Migration cancelled"; \
	fi

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

# ── Neon ────────────────────────────────────────────────────────────────────

# Clean Neon databases (drop all tables and recreate schema)
clean-neon:
	@echo "🧹 Clean Neon databases"
	@echo ""
	@echo "⚠️  WARNING: This will DELETE ALL DATA in your Neon databases!"
	@echo "    This action is IRREVERSIBLE!"
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@read -p "Are you ABSOLUTELY SURE? Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "Cleaning CMS database..."; \
		psql "$(NEON_CMS_DB_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO neondb_owner; GRANT ALL ON SCHEMA public TO public;"; \
		echo "Cleaning Ecommerce database..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO neondb_owner; GRANT ALL ON SCHEMA public TO public;"; \
		echo "✅ Neon databases cleaned!"; \
	else \
		echo "❌ Clean cancelled"; \
	fi

# Export local databases to Neon (schema + data)
export-to-neon:
	@echo "📤 Exporting local databases to Neon..."
	@echo "⚠️  WARNING: This will overwrite existing data in Neon!"
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Dumping CMS database..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_cms_backend --clean --if-exists > cms_dump.sql; \
		echo "Dumping Ecommerce database..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_ecommerce --clean --if-exists > ecommerce_dump.sql; \
		echo "Restoring CMS to Neon..."; \
		psql "$(NEON_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_dump.sql; \
		echo "Restoring Ecommerce to Neon..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_dump.sql; \
		echo "✅ Export complete!"; \
		rm cms_dump.sql ecommerce_dump.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

# Export only data to Neon (assumes schemas already exist)
export-data-only:
	@echo "📤 Exporting only data to Neon..."
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Dumping CMS data..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_cms_backend --data-only --inserts --rows-per-insert=100 > cms_data.sql; \
		echo "Dumping Ecommerce data..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_ecommerce --data-only --inserts --rows-per-insert=100 > ecommerce_data.sql; \
		echo "Restoring CMS data to Neon..."; \
		psql "$(NEON_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_data.sql 2>&1 | grep -v "ERROR.*permission denied.*RI_ConstraintTrigger" | grep -v "ERROR.*duplicate key"; \
		echo "Disabling Ecommerce triggers temporarily..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses DISABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods DISABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders DISABLE TRIGGER trigger_set_order_number;"; \
		echo "Restoring Ecommerce data to Neon..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_data.sql 2>&1 | grep -v "ERROR.*permission denied.*RI_ConstraintTrigger" | grep -v "ERROR.*duplicate key"; \
		echo "Re-enabling Ecommerce triggers..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses ENABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods ENABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders ENABLE TRIGGER trigger_set_order_number;"; \
		echo "✅ Data export complete!"; \
		rm cms_data.sql ecommerce_data.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

# ── Railway ──────────────────────────────────────────────────────────────────

# Export local databases to Railway (schema + data) - use for initial migration
export-to-railway:
	@echo "📤 Exporting local databases to Railway..."
	@echo "⚠️  WARNING: This will overwrite existing data in Railway!"
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Dumping CMS database..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_cms_backend --clean --if-exists > cms_dump.sql; \
		echo "Dumping Ecommerce database..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_ecommerce --clean --if-exists > ecommerce_dump.sql; \
		echo "Restoring CMS to Railway..."; \
		psql "$(RAILWAY_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_dump.sql; \
		echo "Restoring Ecommerce to Railway..."; \
		psql "$(RAILWAY_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_dump.sql; \
		echo "✅ Export to Railway complete!"; \
		rm cms_dump.sql ecommerce_dump.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

# Export only data to Railway (assumes schemas already exist)
export-data-only-railway:
	@echo "📤 Exporting only data to Railway..."
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Dumping CMS data..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_cms_backend --data-only --inserts --rows-per-insert=100 > cms_data.sql; \
		echo "Dumping Ecommerce data..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_ecommerce --data-only --inserts --rows-per-insert=100 > ecommerce_data.sql; \
		echo "Restoring CMS data to Railway..."; \
		psql "$(RAILWAY_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_data.sql 2>&1 | grep -v "ERROR.*duplicate key"; \
		echo "Disabling Ecommerce triggers temporarily..."; \
		psql "$(RAILWAY_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses DISABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods DISABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders DISABLE TRIGGER trigger_set_order_number;"; \
		echo "Restoring Ecommerce data to Railway..."; \
		psql "$(RAILWAY_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_data.sql 2>&1 | grep -v "ERROR.*duplicate key"; \
		echo "Re-enabling Ecommerce triggers..."; \
		psql "$(RAILWAY_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses ENABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods ENABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders ENABLE TRIGGER trigger_set_order_number;"; \
		echo "✅ Data export to Railway complete!"; \
		rm cms_data.sql ecommerce_data.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

# Database shell (Railway CMS)
db-railway-cms:
	@psql "$(RAILWAY_CMS_DB_URL)"

# Database shell (Railway Ecommerce)
db-railway-ecommerce:
	@psql "$(RAILWAY_ECOMMERCE_DB_URL)"

# ── Local ────────────────────────────────────────────────────────────────────

# Clean everything (local)
clean:
	@echo "⚠️  This will remove all local data!"
	@read -p "Are you sure? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		docker-compose down -v; \
		echo "✅ Cleaned"; \
	fi

# Database shell (local)
db:
	@docker-compose exec postgres psql -U postgres

# Database shell (Neon CMS)
db-neon-cms:
	@if [ -z "$(NEON_CMS_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL not set in .env"; \
		exit 1; \
	fi
	@psql "$(NEON_CMS_DB_URL)"

# Database shell (Neon Ecommerce)
db-neon-ecommerce:
	@if [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@psql "$(NEON_ECOMMERCE_DB_URL)"

help:
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║          Modeva Backend - Hybrid Docker Setup                  ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🚀 Local Development:"
	@echo "   make dev                   - Start everything (instant hot-reload!)"
	@echo "   make services              - Start only Docker services"
	@echo "   make stop                  - Stop all services"
	@echo "   make restart               - Restart services"
	@echo "   make logs                  - View Docker logs"
	@echo "   make db                    - PostgreSQL shell (local)"
	@echo ""
	@echo "📦 Migrations:"
	@echo "   make migrate               - Run migrations (local)"
	@echo "   make migrate-railway       - Run migrations on Railway (production)"
	@echo "   make migrate-neon          - Run migrations on Neon (old production)"
	@echo "   make migrate-create        - Create new migration"
	@echo ""
	@echo "🚂 Railway (production):"
	@echo "   make export-to-railway     - Export local DB to Railway (schema + data)"
	@echo "   make export-data-only-railway - Export only data to Railway"
	@echo "   make db-railway-cms        - PostgreSQL shell (Railway CMS)"
	@echo "   make db-railway-ecommerce  - PostgreSQL shell (Railway Ecommerce)"
	@echo "   make migrate-railway       - Run migrations on Railway"
	@echo ""
	@echo "📤 Neon (old):"
	@echo "   make clean-neon            - ⚠️  DROP all tables in Neon (DANGEROUS!)"
	@echo "   make export-to-neon        - Export local DB to Neon (schema + data)"
	@echo "   make export-data-only      - Export only data to Neon"
	@echo "   make db-neon-cms           - PostgreSQL shell (Neon CMS)"
	@echo "   make db-neon-ecommerce     - PostgreSQL shell (Neon Ecommerce)"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "   make clean                 - Remove local Docker volumes"
	@echo ""
	@echo "💡 Recommended workflow to migrate to Railway:"
	@echo "   1. make migrate-railway         # Create tables on Railway"
	@echo "   2. make export-data-only-railway # Import your data"
	@echo ""

.DEFAULT_GOAL := help