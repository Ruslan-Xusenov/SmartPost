#!/bin/bash
# SmartPost Database Initialization Script

set -e

echo "🗄️ SmartPost DB initialization..."

# Wait for PostgreSQL
until pg_isready -h postgres -U smartpost; do
  echo "⏳ Waiting for PostgreSQL..."
  sleep 2
done

echo "✅ PostgreSQL is ready"

# Run migrations
psql -h postgres -U smartpost -d smartpost -f /app/internal/database/migrations/001_init.sql

echo "✅ Migrations applied successfully"
