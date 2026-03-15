#!/bin/bash
# Start all Go microservices in the background for local development.
# Usage: ./scripts/start-services.sh
# Stop:  ./scripts/start-services.sh stop

set -e

PIDS_DIR=".pids"
mkdir -p "$PIDS_DIR"

if [ "$1" = "stop" ]; then
    echo "Stopping all services..."
    for pidfile in "$PIDS_DIR"/*.pid; do
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            svc=$(basename "$pidfile" .pid)
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid"
                echo "  Stopped $svc (PID $pid)"
            fi
            rm "$pidfile"
        fi
    done
    echo "All services stopped."
    exit 0
fi

echo "Starting all services..."

services=(
    "gateway:8080"
    "payment:8081"
    "merchant:8082"
    "settlement:8083"
    "webhook:8084"
    "exchange:8085"
    "subscription:8086"
    "notification:8087"
    "admin:8088"
)

for entry in "${services[@]}"; do
    svc="${entry%%:*}"
    port="${entry##*:}"

    # Kill existing if running
    if [ -f "$PIDS_DIR/$svc.pid" ]; then
        old_pid=$(cat "$PIDS_DIR/$svc.pid")
        kill "$old_pid" 2>/dev/null || true
    fi

    PORT=$port go run "./services/$svc/cmd/" > "/tmp/openpay-$svc.log" 2>&1 &
    echo $! > "$PIDS_DIR/$svc.pid"
    echo "  Started $svc on :$port (PID $!)"
done

echo ""
echo "All 9 services running. Logs in /tmp/openpay-*.log"
echo "Stop with: ./scripts/start-services.sh stop"
echo ""
echo "Gateway:      http://localhost:8080"
echo "Merchant UI:  http://localhost:3000 (run: pnpm --filter merchant-portal dev)"
echo "Admin UI:     http://localhost:3001 (run: pnpm --filter admin-dashboard dev --port 3001)"
