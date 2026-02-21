#!/bin/bash
set -e

echo "🔧 Initializing databases..."

# Function to initialize database with proper schema setup
init_database() {
    local DB_URL=$1
    local MIGRATION_PATH=$2
    local DB_NAME=$3
    
    echo "📦 Initializing $DB_NAME..."
    
    # Create schema_migrations table if it doesn't exist
    psql "$DB_URL" -c "CREATE TABLE IF NOT EXISTS public.schema_migrations (version bigint not null primary key, dirty boolean not null);" 2>/dev/null || true
    
    # Create uuid-ossp extension with explicit schema
    psql "$DB_URL" -c 'SET search_path TO public; CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;' 2>/dev/null || true
    
    echo "🔄 Running $DB_NAME migrations..."
    migrate -path "$MIGRATION_PATH" -database "$DB_URL" up || echo "  ⚠️  $DB_NAME: Migration issue (may already be up-to-date)"
}

# Initialize CMS database
init_database "${CMS_DB_URL}" "/app/migrations/cms" "CMS"

# Initialize Ecommerce database
init_database "${ECOMMERCE_DB_URL}" "/app/migrations/ecommerce" "Ecommerce"

echo "✅ Database initialization complete"
echo "🚀 Starting application..."
exec air -c .air.toml