#!/bin/bash
# =============================================================================
# Open Lanka Payment — Development Environment Starter
# Starts all infrastructure, Go services (with hot reload), and frontends.
#
# Usage:
#   ./scripts/start-dev.sh          Start everything
#   ./scripts/start-dev.sh stop     Stop everything
#   ./scripts/start-dev.sh status   Show status of all services
# =============================================================================

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

# Ensure GOPATH/bin is in PATH (where `go install` puts binaries)
GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
export PATH="$GOBIN:$PATH"

PIDS_DIR="$ROOT_DIR/.pids"
LOGS_DIR="$ROOT_DIR/.logs"
mkdir -p "$PIDS_DIR" "$LOGS_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# Services definition: name:port
GO_SERVICES=(
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

FRONTEND_APPS=(
    "merchant-portal:4600"
    "admin-dashboard:4500"
)

print_banner() {
    echo ""
    echo -e "${BOLD}${BLUE}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${BLUE}║          Open Lanka Payment — Dev Mode           ║${NC}"
    echo -e "${BOLD}${BLUE}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_summary() {
    echo ""
    echo -e "${BOLD}${GREEN}═══════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${GREEN}  All services are running!${NC}"
    echo -e "${BOLD}${GREEN}═══════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${BOLD}  Infrastructure:${NC}"
    echo -e "  ${CYAN}PostgreSQL${NC}       http://localhost:5433"
    echo -e "  ${CYAN}Redis${NC}            localhost:6379"
    echo -e "  ${CYAN}NATS${NC}             nats://localhost:4222"
    echo -e "  ${CYAN}NATS Monitor${NC}     http://localhost:8222"
    echo -e "  ${CYAN}MinIO Console${NC}    http://localhost:9001  ${DIM}(minioadmin / minioadmin123)${NC}"
    echo -e "  ${CYAN}MinIO API${NC}        http://localhost:9000"
    echo -e "  ${CYAN}Mailpit UI${NC}       http://localhost:8025  ${DIM}(catches all dev emails)${NC}"
    echo -e "  ${CYAN}Mailpit SMTP${NC}     localhost:1025"
    echo ""
    echo -e "${BOLD}  Go Services (hot reload via air):${NC}"
    for entry in "${GO_SERVICES[@]}"; do
        svc="${entry%%:*}"
        port="${entry##*:}"
        echo -e "  ${MAGENTA}${svc}${NC}$(printf '%*s' $((18 - ${#svc})) '')http://localhost:${port}"
    done
    echo ""
    echo -e "${BOLD}  Frontend Apps (Vite HMR):${NC}"
    echo -e "  ${YELLOW}Merchant Portal${NC}  http://localhost:4600"
    echo -e "  ${YELLOW}Admin Dashboard${NC}  http://localhost:4500"
    echo ""
    echo -e "${BOLD}  Logs:${NC}"
    echo -e "  ${DIM}Go services:   .logs/<service>.log${NC}"
    echo -e "  ${DIM}Frontends:     .logs/<app>.log${NC}"
    echo ""
    echo -e "${BOLD}  Commands:${NC}"
    echo -e "  ${DIM}Stop all:      make stop${NC}"
    echo -e "  ${DIM}Status:        make status${NC}"
    echo -e "  ${DIM}Logs:          tail -f .logs/<service>.log${NC}"
    echo ""
}

check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: '$1' is not installed.${NC}"
        echo -e "${DIM}$2${NC}"
        return 1
    fi
}

stop_all() {
    echo -e "${YELLOW}Stopping all services...${NC}"

    # Stop Go services
    for entry in "${GO_SERVICES[@]}"; do
        svc="${entry%%:*}"
        pidfile="$PIDS_DIR/$svc.pid"
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null || true
                echo -e "  ${RED}Stopped${NC} $svc (PID $pid)"
            fi
            rm -f "$pidfile"
        fi
    done

    # Stop frontend apps
    for entry in "${FRONTEND_APPS[@]}"; do
        app="${entry%%:*}"
        pidfile="$PIDS_DIR/$app.pid"
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null || true
                echo -e "  ${RED}Stopped${NC} $app (PID $pid)"
            fi
            rm -f "$pidfile"
        fi
    done

    # Stop docker infrastructure
    echo -e "  ${YELLOW}Stopping Docker containers...${NC}"
    docker compose down 2>/dev/null || true

    echo -e "${GREEN}All services stopped.${NC}"
    exit 0
}

show_status() {
    echo ""
    echo -e "${BOLD}Service Status:${NC}"
    echo ""

    # Docker containers
    echo -e "${BOLD}  Infrastructure (Docker):${NC}"
    for container in postgres redis nats minio mailpit; do
        if docker compose ps --status running 2>/dev/null | grep -q "$container"; then
            echo -e "  ${GREEN}●${NC} $container"
        else
            echo -e "  ${RED}●${NC} $container"
        fi
    done

    echo ""
    echo -e "${BOLD}  Go Services:${NC}"
    for entry in "${GO_SERVICES[@]}"; do
        svc="${entry%%:*}"
        port="${entry##*:}"
        pidfile="$PIDS_DIR/$svc.pid"
        if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
            echo -e "  ${GREEN}●${NC} $svc :$port"
        else
            echo -e "  ${RED}●${NC} $svc :$port"
        fi
    done

    echo ""
    echo -e "${BOLD}  Frontend Apps:${NC}"
    for entry in "${FRONTEND_APPS[@]}"; do
        app="${entry%%:*}"
        port="${entry##*:}"
        pidfile="$PIDS_DIR/$app.pid"
        if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
            echo -e "  ${GREEN}●${NC} $app :$port"
        else
            echo -e "  ${RED}●${NC} $app :$port"
        fi
    done
    echo ""
    exit 0
}

# Handle commands
case "${1:-}" in
    stop)  stop_all ;;
    status) show_status ;;
esac

print_banner

# ─── Check prerequisites ───
echo -e "${BOLD}Checking prerequisites...${NC}"
MISSING=0
check_tool "docker" "Install Docker: https://docs.docker.com/get-docker/" || MISSING=1
check_tool "go" "Install Go: https://go.dev/dl/" || MISSING=1
check_tool "air" "Install air: go install github.com/air-verse/air@latest" || MISSING=1
check_tool "pnpm" "Install pnpm: npm install -g pnpm" || MISSING=1
if [ "$MISSING" -eq 1 ]; then
    echo -e "${RED}Please install missing tools and try again.${NC}"
    exit 1
fi
echo -e "${GREEN}All prerequisites found.${NC}"
echo ""

# ─── Copy .env if missing ───
if [ ! -f "$ROOT_DIR/.env" ]; then
    echo -e "${YELLOW}No .env file found — copying from .env.example${NC}"
    cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
fi

# Source .env
set -a
source "$ROOT_DIR/.env"
set +a

# ─── Start Infrastructure ───
echo -e "${BOLD}Starting infrastructure...${NC}"
docker compose up -d postgres redis nats minio minio-init mailpit

echo -e "  Waiting for PostgreSQL..."
TRIES=0
until docker compose exec -T postgres pg_isready -U olp 2>/dev/null; do
    TRIES=$((TRIES + 1))
    if [ "$TRIES" -ge 30 ]; then
        echo -e "${RED}PostgreSQL failed to start!${NC}"
        exit 1
    fi
    sleep 1
done
echo -e "  ${GREEN}PostgreSQL ready.${NC}"

echo -e "  Waiting for other services..."
sleep 2
echo -e "  ${GREEN}Infrastructure ready.${NC}"
echo ""

# ─── Run Migrations ───
if command -v migrate &> /dev/null; then
    echo -e "${BOLD}Running database migrations...${NC}"
    DB_URL_BASE="postgres://olp:olp_dev_password@localhost:5433"
    for db in merchant payment settlement exchange webhook subscription admin notification; do
        migrate -path "migrations/$db" -database "${DB_URL_BASE}/${db}_db?sslmode=disable" up 2>/dev/null || true
    done
    echo -e "  ${GREEN}Migrations complete.${NC}"
    echo ""
else
    echo -e "${YELLOW}Skipping migrations — 'migrate' CLI not found.${NC}"
    echo -e "${DIM}  Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest${NC}"
    echo ""
fi

# ─── Start Go Services with Air (hot reload) ───
echo -e "${BOLD}Starting Go services with hot reload...${NC}"
for entry in "${GO_SERVICES[@]}"; do
    svc="${entry%%:*}"
    port="${entry##*:}"

    # Kill existing if running
    if [ -f "$PIDS_DIR/$svc.pid" ]; then
        old_pid=$(cat "$PIDS_DIR/$svc.pid")
        kill "$old_pid" 2>/dev/null || true
        rm -f "$PIDS_DIR/$svc.pid"
    fi

    # Start with air for hot reload
    PORT=$port air \
        --build.cmd "go build -o ./tmp/air-$svc ./services/$svc/cmd/" \
        --build.bin "./tmp/air-$svc" \
        --build.include_ext "go" \
        --build.exclude_dir "bin,tmp,node_modules,dist,apps,migrations,scripts,config,e2e,.git,.pids,.logs" \
        > "$LOGS_DIR/$svc.log" 2>&1 &

    echo $! > "$PIDS_DIR/$svc.pid"
    echo -e "  ${GREEN}Started${NC} $svc on :$port (PID $!)"
done
echo ""

# ─── Start Frontend Apps ───
echo -e "${BOLD}Starting frontend apps with Vite HMR...${NC}"
for entry in "${FRONTEND_APPS[@]}"; do
    app="${entry%%:*}"
    port="${entry##*:}"

    # Kill existing if running
    if [ -f "$PIDS_DIR/$app.pid" ]; then
        old_pid=$(cat "$PIDS_DIR/$app.pid")
        kill "$old_pid" 2>/dev/null || true
        rm -f "$PIDS_DIR/$app.pid"
    fi

    # Install deps if needed
    if [ ! -d "$ROOT_DIR/apps/$app/node_modules" ]; then
        echo -e "  ${YELLOW}Installing deps for $app...${NC}"
        (cd "$ROOT_DIR/apps/$app" && pnpm install) > "$LOGS_DIR/$app-install.log" 2>&1
    fi

    # Start Vite dev server
    (cd "$ROOT_DIR/apps/$app" && pnpm dev) > "$LOGS_DIR/$app.log" 2>&1 &
    echo $! > "$PIDS_DIR/$app.pid"
    echo -e "  ${GREEN}Started${NC} $app on :$port (PID $!)"
done

print_summary
