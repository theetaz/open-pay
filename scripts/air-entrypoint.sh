#!/bin/sh
# Per-service air launcher for containerized hot reload.
# Generates an air config from the $SERVICE env var, then watches that service
# (plus the shared pkg/) and rebuilds on change. Polling is enabled so file
# events propagate reliably across bind mounts on macOS/Windows Docker.
set -e

: "${SERVICE:?SERVICE env var is required (e.g. merchant)}"
: "${PORT:=8080}"
export PORT

CONFIG="/tmp/air-${SERVICE}.toml"

cat > "$CONFIG" <<EOF
root = "/app"
tmp_dir = "/tmp/air"

[build]
cmd = "go build -o /tmp/air/${SERVICE} ./services/${SERVICE}/cmd/"
bin = "/tmp/air/${SERVICE}"
include_ext = ["go"]
include_dir = ["services/${SERVICE}", "pkg"]
poll = true
poll_interval = 500
delay = 200
stop_on_error = true
EOF

cd /app
exec air -c "$CONFIG"
