#!/bin/bash
set -e

cd /opt/openpay

echo "==> Pulling latest code..."
git pull origin main

echo "==> Building service images..."
docker compose -f docker-compose.prod.yml --env-file .env.prod build --parallel

echo "==> Running migrations..."
for db in merchant payment settlement exchange webhook subscription admin notification directdebit; do
    echo "Migrating ${db}..."
    migrate -path migrations/${db} -database "postgres://olp:${POSTGRES_PASSWORD:-olp_prod_Str0ngP4ss2026}@localhost:5432/${db}_db?sslmode=disable" up 2>&1 || true
done

echo "==> Restarting services..."
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

echo "==> Pruning old images..."
docker image prune -f

echo "==> Deploy complete!"
docker compose -f docker-compose.prod.yml --env-file .env.prod ps --format "table {{.Name}}\t{{.Status}}"
