#!/bin/bash
set -e

echo "🧪 BetKZ E2E Test Runner"
echo "========================"

# Start infrastructure
echo "📦 Starting PostgreSQL and Redis..."
cd "$(dirname "$0")/../deployments"

docker compose up -d db redis
echo "⏳ Waiting for databases to be healthy..."
sleep 5

# Create test database
echo "📊 Creating test database..."
docker exec betkz-db psql -U betkz -c "DROP DATABASE IF EXISTS betkz_test" 2>/dev/null || true
docker exec betkz-db psql -U betkz -c "CREATE DATABASE betkz_test" 2>/dev/null || true

# Run E2E tests
echo "🚀 Running E2E tests..."
cd "$(dirname "$0")/../backend"

E2E_TESTS=1 \
TEST_DATABASE_URL="postgres://betkz:betkz_dev_pass@localhost:5432/betkz_test?sslmode=disable" \
TEST_REDIS_URL="redis://localhost:6379/1" \
go test -v -count=1 -timeout 120s ./cmd/api/... 2>&1

EXIT_CODE=$?

# Cleanup test database
echo "🧹 Cleaning up test database..."
docker exec betkz-db psql -U betkz -c "DROP DATABASE IF EXISTS betkz_test" 2>/dev/null || true

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ All E2E tests passed!"
else
    echo "❌ Some tests failed (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
