ifneq ($(wildcard .env),)
include .env
export
endif

DEV_ENV_FILE ?= .env.dev
PROD_ENV_FILE ?= .env.prod

DEV_STACK = docker compose --env-file $(DEV_ENV_FILE) -f docker-compose.dev.yml
PROD_STACK = docker compose --env-file $(PROD_ENV_FILE) -f docker-compose.prod.yml
INTEGRATION_TEST_STACK = $(DEV_STACK) -f docker-compose-integration-test.yml
ALL_STACK = $(INTEGRATION_TEST_STACK)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

ensure-dev-env: ## Create .env.dev from example if it does not exist
	@if [ ! -f "$(DEV_ENV_FILE)" ]; then \
		cp .env.dev.example "$(DEV_ENV_FILE)"; \
		echo "Created $(DEV_ENV_FILE) from .env.dev.example"; \
	fi
.PHONY: ensure-dev-env

ensure-prod-env: ## Create .env.prod from example if it does not exist
	@if [ ! -f "$(PROD_ENV_FILE)" ]; then \
		cp .env.prod.example "$(PROD_ENV_FILE)"; \
		echo "Created $(PROD_ENV_FILE) from .env.prod.example"; \
	fi
.PHONY: ensure-prod-env

preflight-dev: ensure-dev-env ## Validate required env vars for dev profile
	./scripts/preflight-env.sh $(DEV_ENV_FILE) dev
.PHONY: preflight-dev

preflight-prod: ensure-prod-env ## Validate required env vars for prod profile
	./scripts/preflight-env.sh $(PROD_ENV_FILE) prod
.PHONY: preflight-prod

compose-up: preflight-dev ### Run docker compose (without backend and reverse proxy)
	$(DEV_STACK) up --build -d db keycloak && docker compose logs -f
.PHONY: compose-up

compose-up-all: preflight-dev ### Run docker compose (with backend and reverse proxy)
	$(DEV_STACK) up --build -d
.PHONY: compose-up-all

dev-up: preflight-dev ## Start the complete development stack
	$(DEV_STACK) up --build -d db keycloak keycloak-db
	@until curl --silent --show-error --fail http://127.0.0.1:$${DEV_KEYCLOAK_HOST_PORT:-8082}/realms/lc/.well-known/openid-configuration >/dev/null; do \
		sleep 1; \
	done
	$(DEV_STACK) up --build -d app frontend-next frontend-docs nginx
.PHONY: dev-up

compose-up-prod: preflight-prod ### Run production docker compose stack
	$(PROD_STACK) up --build -d
.PHONY: compose-up-prod

compose-up-integration-test: preflight-dev ### Run integration test container
	$(INTEGRATION_TEST_STACK) up --build --abort-on-container-exit --exit-code-from integration-test; \
	exit_code=$$?; \
	$(INTEGRATION_TEST_STACK) down --remove-orphans; \
	exit $$exit_code
.PHONY: compose-up-integration-test

compose-down: ### Down docker compose
	$(ALL_STACK) down --remove-orphans
.PHONY: compose-down

dev-down: ## Stop the complete development stack
	$(DEV_STACK) down --remove-orphans
.PHONY: dev-down

compose-down-prod: ### Down production docker compose stack
	$(PROD_STACK) down --remove-orphans
.PHONY: compose-down-prod

swag-v1: ### swag init
	swag init --parseDependency -g internal/app/controller/restapi/router.go
.PHONY: swag-v1

deps: ### deps tidy + verify
	go mod tidy && go mod verify
.PHONY: deps

deps-audit: ### check dependencies vulnerabilities
	govulncheck ./...
.PHONY: deps-audit

fix-diff: ### Show code changes by `go fix`
	go fix -diff ./...
.PHONY: fix-diff

format: ### Run code formatter
	go fix ./...
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
.PHONY: format

run-app: deps swag-v1 ### run API app
	go mod download && \
	CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run-app

run-worker: deps ### run background workers
	go mod download && \
	CGO_ENABLED=0 go run ./cmd/worker
.PHONY: run-worker

build-worker: deps ### build worker binary
	go mod download && \
	CGO_ENABLED=0 go build -o ./bin/worker ./cmd/worker
.PHONY: build-worker


run: deps swag-v1 build-worker ### run app + worker
	@set -e; \
	trap 'status=$$?; if [ -n "$$worker_pid" ]; then kill $$worker_pid 2>/dev/null || true; wait $$worker_pid 2>/dev/null || true; fi; exit $$status' INT TERM EXIT; \
	go mod download; \
	# Build worker binary locally and run it in background
	mkdir -p ./bin; \
	./bin/worker & \
	worker_pid=$$!; \
	# Run API in foreground
	CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

docker-rm-volume: ### remove docker volume
	docker volume rm backend_db_data
.PHONY: docker-rm-volume

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### check by hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ### check by dotenv linter
	dotenv-linter
.PHONY: linter-dotenv

test: ### run test
	go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/... ./pkg/...
.PHONY: test

integration-test: ### run integration tests against the current stack
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

mock: ### run mockgen
	mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
	mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go
.PHONY: mock

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations '$(word 2,$(MAKECMDGOALS))'
.PHONY: migrate-create

migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up
.PHONY: migrate-up

bin-deps: ### install tools
	go install tool
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
.PHONY: bin-deps

pre-commit: deps swag-v1 proto-v1 mock format linter-golangci test ### run pre-commit
.PHONY: pre-commit
