#!/bin/bash
set -e

# =============================================================================
# Open Pay — Environment-Aware Deploy Script
# Usage: ./deploy.sh <environment>
#   environment: staging | production
# =============================================================================

ENVIRONMENT="${1:-staging}"

case "$ENVIRONMENT" in
  staging)
    BRANCH="develop"
    ENV_FILE=".env.staging"
    ;;
  production)
    BRANCH="main"
    ENV_FILE=".env.prod"
    ;;
  *)
    echo "ERROR: Unknown environment '$ENVIRONMENT'. Use 'staging' or 'production'."
    exit 1
    ;;
esac

echo "============================================"
echo "  Deploying Open Pay — $ENVIRONMENT"
echo "  Branch: $BRANCH"
echo "  Env file: $ENV_FILE"
echo "============================================"

cd /opt/openpay

# --- Validate env file exists ---
if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: Environment file '$ENV_FILE' not found."
  echo "Copy the example and fill in values:"
  echo "  cp ${ENV_FILE}.example $ENV_FILE"
  exit 1
fi

# --- Pull latest code ---
echo "==> Fetching latest code from $BRANCH..."
git fetch origin "$BRANCH"
git reset --hard "origin/$BRANCH"

# --- Load env for migrations ---
set -a
source "$ENV_FILE"
set +a

# --- Ensure infrastructure is running ---
echo "==> Ensuring infrastructure is running..."
docker compose -f docker-compose.yml up -d postgres redis nats minio minio-init mailpit
echo "Waiting for Postgres..."
until docker exec openpay-postgres-1 pg_isready -U olp >/dev/null 2>&1; do
  echo "  ...waiting"
  sleep 1
done
echo "Postgres ready."

# --- Build service images ---
echo "==> Building service images..."
docker compose -f docker-compose.prod.yml --env-file "$ENV_FILE" build --parallel

# --- Run migrations ---
echo "==> Running migrations..."
MIGRATE_DB_HOST="${MIGRATE_DB_HOST:-localhost}"
MIGRATE_DB_PORT="${MIGRATE_DB_PORT:-5432}"
for db in merchant payment settlement exchange webhook subscription admin notification directdebit; do
  echo "Migrating ${db}..."
  migrate -path "migrations/${db}" \
    -database "postgres://olp:${POSTGRES_PASSWORD}@${MIGRATE_DB_HOST}:${MIGRATE_DB_PORT}/${db}_db?sslmode=disable" \
    up 2>&1 || true
done

# --- Start services ---
echo "==> Starting services (recreating containers)..."
docker compose -f docker-compose.prod.yml --env-file "$ENV_FILE" up -d --force-recreate

# --- Cleanup ---
echo "==> Pruning old images..."
docker image prune -f

# --- Health check ---
echo "==> Checking service health..."
sleep 10
docker compose -f docker-compose.prod.yml --env-file "$ENV_FILE" ps --format "table {{.Name}}\t{{.Status}}"

echo "============================================"
echo "  Deploy complete — $ENVIRONMENT"
echo "============================================"
