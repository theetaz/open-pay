#!/bin/sh
# Runs golang-migrate against every service database from inside a container.
# Invoked by the `migrate` service in docker-compose.dev.yml.
#
# Usage (via entrypoint arg):
#   up         -> apply all migrations (default)
#   down       -> roll back one step in every database
#   drop       -> drop everything in every database (DESTRUCTIVE)
set -e

ACTION="${1:-up}"
DB_BASE="postgres://olp:${POSTGRES_PASSWORD:-olp_dev_password}@postgres:5432"
DBS="merchant payment settlement exchange webhook subscription admin notification directdebit"

for db in $DBS; do
  url="${DB_BASE}/${db}_db?sslmode=disable"
  case "$ACTION" in
    up)
      echo "Migrating ${db}_db..."
      migrate -path "/migrations/$db" -database "$url" up
      ;;
    down)
      echo "Rolling back ${db}_db..."
      migrate -path "/migrations/$db" -database "$url" down 1
      ;;
    drop)
      echo "Dropping ${db}_db..."
      migrate -path "/migrations/$db" -database "$url" drop -f
      ;;
    *)
      echo "Unknown action: $ACTION (expected up|down|drop)" >&2
      exit 1
      ;;
  esac
done

echo "Migration '$ACTION' complete for all databases."
