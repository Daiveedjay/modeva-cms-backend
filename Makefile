# ════════════════════════════════════════════════════════════
# Modeva Backend - Hybrid Docker Setup
# PostgreSQL & Redis in Docker, Air on Windows
# ════════════════════════════════════════════════════════════

.PHONY: dev services stop restart logs migrate migrate-create migrate-neon clean help export-to-neon export-data-only clean-neon

# Load environment variables from .env (includes Neon credentials)
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

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
	@echo "⚠️  This will run migrations on your PRODUCTION databases!"
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		echo "   Please add these variables to your .env file"; \
		exit 1; \
	fi
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Initializing CMS migration system..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database '$(NEON_CMS_DB_URL)' force 0" 2>/dev/null || true; \
		echo "Running CMS migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/cms -database '$(NEON_CMS_DB_URL)' up" || echo "  ⚠️  CMS: Migration issue (may already be up-to-date)"; \
		echo "Initializing Ecommerce migration system..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database '$(NEON_ECOMMERCE_DB_URL)' force 0" 2>/dev/null || true; \
		echo "Running Ecommerce migrations..."; \
		wsl -d Ubuntu bash -lc "cd /mnt/c/Users/Jajad/Desktop/MODEVA/CMS/modeva-cms-backend && migrate -path ./migrations/ecommerce -database '$(NEON_ECOMMERCE_DB_URL)' up" || echo "  ⚠️  Ecommerce: Migration issue (may already be up-to-date)"; \
		echo "✅ Neon migrations complete"; \
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
		echo ""; \
		echo "💡 Next steps:"; \
		echo "   1. Run 'make migrate-neon' to create tables"; \
		echo "   2. Run 'make export-data-only' to import your data"; \
	else \
		echo "❌ Clean cancelled"; \
	fi

# Export local databases to Neon (schema + data) - IMPROVED VERSION
export-to-neon:
	@echo "📤 Exporting local databases to Neon..."
	@echo ""
	@echo "⚠️  WARNING: This will overwrite existing data in Neon!"
	@echo "    Recommended: Run 'make clean-neon' first for a clean import"
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
		echo "✅ Dumps created"; \
		echo ""; \
		echo "Restoring CMS to Neon..."; \
		psql "$(NEON_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_dump.sql; \
		echo ""; \
		echo "Restoring Ecommerce to Neon..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_dump.sql; \
		echo ""; \
		echo "✅ Export complete!"; \
		echo "⚠️  Check output above for any errors"; \
		rm cms_dump.sql ecommerce_dump.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

# Export only data to Neon (assumes schemas already exist) - WITH TRIGGER MANAGEMENT
export-data-only:
	@echo "📤 Exporting only data to Neon..."
	@echo ""
	@echo "⚠️  WARNING: This will add/overwrite data in Neon!"
	@echo "    (Schemas must already exist on Neon)"
	@if [ -z "$(NEON_CMS_DB_URL)" ] || [ -z "$(NEON_ECOMMERCE_DB_URL)" ]; then \
		echo "❌ ERROR: NEON_CMS_DB_URL or NEON_ECOMMERCE_DB_URL not set in .env"; \
		exit 1; \
	fi
	@read -p "Continue? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "Dumping CMS data (using INSERT format)..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_cms_backend --data-only --inserts --rows-per-insert=100 > cms_data.sql; \
		echo "Dumping Ecommerce data (using INSERT format)..."; \
		pg_dump -h localhost -p 5433 -U postgres -d modeva_ecommerce --data-only --inserts --rows-per-insert=100 > ecommerce_data.sql; \
		echo "✅ Data dumps created"; \
		echo ""; \
		echo "Restoring CMS data to Neon..."; \
		psql "$(NEON_CMS_DB_URL)" -v ON_ERROR_STOP=0 < cms_data.sql 2>&1 | grep -v "ERROR.*permission denied.*RI_ConstraintTrigger" | grep -v "ERROR.*duplicate key"; \
		echo ""; \
		echo "Disabling Ecommerce triggers temporarily..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses DISABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods DISABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders DISABLE TRIGGER trigger_set_order_number;"; \
		echo "Restoring Ecommerce data to Neon..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -v ON_ERROR_STOP=0 < ecommerce_data.sql 2>&1 | grep -v "ERROR.*permission denied.*RI_ConstraintTrigger" | grep -v "ERROR.*duplicate key"; \
		echo "Re-enabling Ecommerce triggers..."; \
		psql "$(NEON_ECOMMERCE_DB_URL)" -c "ALTER TABLE addresses ENABLE TRIGGER ensure_single_default_address_trigger; ALTER TABLE user_payment_methods ENABLE TRIGGER trigger_single_default_payment_method; ALTER TABLE orders ENABLE TRIGGER trigger_set_order_number;"; \
		echo ""; \
		echo "✅ Data export complete!"; \
		echo "⚠️  Check output above for any real errors"; \
		rm cms_data.sql ecommerce_data.sql; \
	else \
		echo "❌ Export cancelled"; \
	fi

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
	@echo "   make dev              - Start everything (instant hot-reload!)"
	@echo "   make services         - Start only Docker services"
	@echo "   make stop             - Stop all services"
	@echo "   make restart          - Restart services"
	@echo "   make logs             - View Docker logs"
	@echo "   make db               - PostgreSQL shell (local)"
	@echo ""
	@echo "📦 Migrations:"
	@echo "   make migrate          - Run migrations (local)"
	@echo "   make migrate-neon     - Run migrations on Neon (production)"
	@echo "   make migrate-create   - Create new migration"
	@echo ""
	@echo "📤 Neon Export/Import:"
	@echo "   make clean-neon       - ⚠️  DROP all tables in Neon (DANGEROUS!)"
	@echo "   make export-to-neon   - Export local DB to Neon (schema + data)"
	@echo "   make export-data-only - Export only data to Neon (no schema)"
	@echo "   make db-neon-cms      - PostgreSQL shell (Neon CMS)"
	@echo "   make db-neon-ecommerce - PostgreSQL shell (Neon Ecommerce)"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "   make clean            - Remove local Docker volumes"
	@echo ""
	@echo "💡 Recommended workflow for fresh Neon setup:"
	@echo "   1. make clean-neon          # Clean Neon databases"
	@echo "   2. make migrate-neon        # Run migrations on Neon"
	@echo "   3. make export-data-only    # Import your data"
	@echo ""
	@echo "⚙️  Configuration:"
	@echo "   Neon credentials are loaded from .env file"
	@echo "   Required variables: NEON_CMS_DB_URL, NEON_ECOMMERCE_DB_URL"

.DEFAULT_GOAL := help