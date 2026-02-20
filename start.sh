#!/bin/bash
set -e

echo "🔄 Running CMS migrations..."
migrate -path /app/migrations/cms -database "${CMS_DB_URL}" up

echo "🔄 Running Ecommerce migrations..."
migrate -path /app/migrations/ecommerce -database "${ECOMMERCE_DB_URL}" up

echo "🚀 Starting application..."
exec air -c .air.toml