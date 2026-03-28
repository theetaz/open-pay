#!/bin/bash
set -e

cd /opt/openpay

echo "==> Pulling latest code..."
git pull origin main

echo "==> Ensuring infrastructure is running..."
docker compose -f docker-compose.yml up -d postgres redis nats minio minio-init mailpit
echo "Waiting for Postgres..."
until docker exec openpay-postgres-1 pg_isready -U olp >/dev/null 2>&1; do sleep 1; done
echo "Postgres ready."

echo "==> Building service images..."
docker compose -f docker-compose.prod.yml --env-file .env.prod build --parallel

echo "==> Running migrations..."
for db in merchant payment settlement exchange webhook subscription admin notification directdebit; do
    echo "Migrating ${db}..."
    migrate -path migrations/${db} -database "postgres://olp:${POSTGRES_PASSWORD:-olp_prod_Str0ngP4ss2026}@localhost:5432/${db}_db?sslmode=disable" up 2>&1 || true
done

echo "==> Starting services (recreating only changed containers)..."
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --force-recreate

echo "==> Pruning old images..."
docker image prune -f

echo "==> Waiting for services to be healthy..."
sleep 10

echo "==> Deploy complete!"
docker compose -f docker-compose.prod.yml --env-file .env.prod ps --format "table {{.Name}}\t{{.Status}}"
