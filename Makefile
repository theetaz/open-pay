.PHONY: help start stop restart status logs dev up down down-clean up-observability build test test-v test-integration test-all test-coverage lint fmt vet generate migrate migrate-down migrate-create db-reset clean tidy docker-build docker-push load-test start-native stop-native status-native

# Ignore any COMPOSE_FILE exported by the user's shell — our -f flags are authoritative.
unexport COMPOSE_FILE

# docker compose invocation for the full containerized dev stack (infra + apps)
COMPOSE = docker compose -f docker-compose.yml -f docker-compose.dev.yml
INFRA_SERVICES = postgres redis nats minio minio-init mailpit

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Development (fully containerized — only Docker required) ───
start: ## Build & run the entire system in containers with hot reload
	$(COMPOSE) up -d --build
	@echo ""
	@echo "System is up (containers). Useful URLs:"
	@echo "  API Gateway:      http://localhost:7000"
	@echo "  Merchant Portal:  http://localhost:7010"
	@echo "  Admin Dashboard:  http://localhost:7011"
	@echo "  Mailpit UI:       http://localhost:7027"
	@echo "  MinIO Console:    http://localhost:7025  (minioadmin / minioadmin123)"
	@echo ""
	@echo "  Logs:   make logs   (all)  |  make logs svc=merchant"
	@echo "  Status: make status        Stop:  make stop"
	@echo ""

stop: ## Stop & remove all containers
	$(COMPOSE) down

restart: ## Restart a service (usage: make restart svc=merchant)
	$(COMPOSE) restart $(svc)

status: ## Show status of all containers
	$(COMPOSE) ps

logs: ## Tail logs (all, or one service: make logs svc=gateway)
	$(COMPOSE) logs -f $(svc)

dev: up ## Start dev environment (infra only)
	@echo ""
	@echo "Infrastructure ready:"
	@echo "  PostgreSQL:    localhost:7020"
	@echo "  Redis:         localhost:7021"
	@echo "  NATS:          localhost:7022"
	@echo "  NATS Monitor:  http://localhost:7023"
	@echo "  MinIO Console: http://localhost:7025  (minioadmin / minioadmin123)"
	@echo "  MinIO API:     http://localhost:7024"
	@echo "  Mailpit UI:    http://localhost:7027"
	@echo ""

up: ## Start infrastructure containers only
	$(COMPOSE) up -d $(INFRA_SERVICES)
	@echo "Infrastructure is up."

down: ## Stop all containers
	$(COMPOSE) down

down-clean: ## Stop all containers and remove volumes
	$(COMPOSE) down -v

up-observability: ## Start observability stack (Prometheus, Grafana)
	docker compose --profile observability up -d

# ─── Native (host) dev — fallback to the old air/pnpm workflow ───
start-native: ## Start everything as host processes (requires go, air, pnpm)
	@./scripts/start-dev.sh

stop-native: ## Stop host processes started by start-native
	@./scripts/start-dev.sh stop

status-native: ## Status of host processes
	@./scripts/start-dev.sh status

# ─── Build ───
build: ## Build all Go services
	@for svc in gateway payment merchant settlement webhook exchange subscription notification admin directdebit; do \
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

# ─── Database (runs in the migrate container — no host tools needed) ───
migrate: ## Run all database migrations
	$(COMPOSE) run --rm migrate up

migrate-down: ## Rollback last migration for all databases
	$(COMPOSE) run --rm migrate down

migrate-create: ## Create migration (usage: make migrate-create svc=payment name=create_payments)
	@mkdir -p migrations/$(svc)
	$(COMPOSE) run --rm --entrypoint migrate migrate \
		create -ext sql -dir /migrations/$(svc) -seq $(name)

db-reset: ## Drop everything and re-run all migrations
	$(COMPOSE) run --rm migrate drop
	$(COMPOSE) run --rm migrate up
	@echo "All databases reset."

# ─── Docker ───
docker-build: ## Build all service Docker images
	@for svc in gateway payment merchant settlement webhook exchange subscription notification admin directdebit; do \
		echo "Building openpay/$$svc..."; \
		docker build --build-arg SERVICE=$$svc -t openpay/$$svc -f Dockerfile.service . ; \
	done
	@echo "All Docker images built."

docker-push: ## Push all service Docker images to registry
	@for svc in gateway payment merchant settlement webhook exchange subscription notification admin directdebit; do \
		echo "Pushing openpay/$$svc..."; \
		docker push openpay/$$svc; \
	done

# ─── Load Testing ───
load-test: ## Run k6 load tests (requires k6 installed)
	@echo "Running exchange rates load test..."
	k6 run tests/load/exchange-rates.js --quiet
	@echo ""
	@echo "Running list payments load test..."
	k6 run tests/load/list-payments.js --quiet
	@echo ""
	@echo "Running create payment load test..."
	k6 run tests/load/create-payment.js --quiet

# ─── Cleanup ───
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

# ─── Go Module ───
tidy: ## Tidy go modules
	go mod tidy
