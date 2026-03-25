.PHONY: help start stop status dev up down build test test-v test-coverage lint fmt generate migrate migrate-down migrate-create clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Development ───
start: ## Start everything (infra + Go services + frontends) with hot reload
	@./scripts/start-dev.sh

stop: ## Stop all running services
	@./scripts/start-dev.sh stop

status: ## Show status of all services
	@./scripts/start-dev.sh status

dev: up ## Start dev environment (infra only)
	@echo ""
	@echo "Infrastructure ready:"
	@echo "  PostgreSQL:    localhost:5433"
	@echo "  Redis:         localhost:6379"
	@echo "  NATS:          localhost:4222"
	@echo "  NATS Monitor:  http://localhost:8222"
	@echo "  MinIO Console: http://localhost:9001  (minioadmin / minioadmin123)"
	@echo "  MinIO API:     http://localhost:9000"
	@echo "  Mailpit UI:    http://localhost:8025"
	@echo "  Mailpit SMTP:  localhost:1025"
	@echo ""

up: ## Start infrastructure containers
	docker compose up -d postgres redis nats minio minio-init mailpit
	@echo "Waiting for services to be healthy..."
	@docker compose exec -T postgres sh -c 'until pg_isready -U olp; do sleep 1; done' 2>/dev/null
	@echo "Infrastructure is up."

down: ## Stop all containers
	docker compose down

down-clean: ## Stop all containers and remove volumes
	docker compose down -v

up-observability: ## Start observability stack (Prometheus, Grafana)
	docker compose --profile observability up -d

# ─── Build ───
build: ## Build all Go services
	@for svc in gateway payment merchant settlement webhook exchange subscription notification admin; do \
		echo "Building $$svc..."; \
		go build -o bin/$$svc ./services/$$svc/cmd/; \
	done
	@echo "All services built."

# ─── Testing ───
test: ## Run all unit tests
	go test -short -count=1 ./pkg/... ./services/...

test-v: ## Run all unit tests (verbose)
	go test -short -count=1 -v ./pkg/... ./services/...

test-integration: ## Run integration tests (requires Docker)
	go test -count=1 -run Integration ./pkg/... ./services/...

test-all: ## Run all tests
	go test -count=1 ./pkg/... ./services/...

test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./pkg/... ./services/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Code Quality ───
lint: ## Run Go linter
	golangci-lint run ./pkg/... ./services/...

fmt: ## Format Go code
	gofmt -s -w pkg/ services/

vet: ## Run go vet
	go vet ./pkg/... ./services/...

# ─── Database ───
DB_URL_BASE=postgres://olp:olp_dev_password@localhost:5433

migrate: ## Run all database migrations
	@for db in merchant payment settlement exchange webhook subscription admin notification; do \
		echo "Migrating $$db..."; \
		migrate -path migrations/$$db -database "$(DB_URL_BASE)/$${db}_db?sslmode=disable" up; \
	done
	@echo "All migrations complete."

migrate-down: ## Rollback last migration for all databases
	@for db in merchant payment settlement exchange webhook subscription admin notification; do \
		echo "Rolling back $$db..."; \
		migrate -path migrations/$$db -database "$(DB_URL_BASE)/$${db}_db?sslmode=disable" down 1; \
	done

migrate-create: ## Create migration (usage: make migrate-create svc=payment name=create_payments)
	@mkdir -p migrations/$(svc)
	migrate create -ext sql -dir migrations/$(svc) -seq $(name)

# ─── Cleanup ───
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

# ─── Go Module ───
tidy: ## Tidy go modules
	go mod tidy
