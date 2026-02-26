BIN        := leaderboard-server
DOCKER_TAG := leaderboard-system:latest

.PHONY: build test test-integration lint fmt proto-gen docker-build run up down reset tidy clean help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  run               Start services, build, and run the server"
	@echo "  up                Start MySQL and Redis (detached)"
	@echo "  down              Stop Docker Compose services (keep data)"
	@echo "  reset             Stop services and remove all volumes"
	@echo "  build             Build static binary (CGO_ENABLED=0)"
	@echo "  test              Unit tests with race detector and coverage"
	@echo "  test-integration  Integration tests (requires Docker)"
	@echo "  fmt               Auto-format Go source files (gofmt)"
	@echo "  lint              Run golangci-lint"
	@echo "  proto-gen         Regenerate protobuf code (requires buf)"
	@echo "  docker-build      Build Docker image"
	@echo "  tidy              Tidy go.mod and go.sum"
	@echo "  clean             Remove binary, test cache, and coverage files"

up:
	docker compose up -d

down:
	docker compose kill && docker compose rm -f

reset:
	docker compose kill && docker compose rm -f
	docker volume rm leaderboard_system_mysql_data leaderboard_system_redis_data 2>/dev/null || true

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server/

test:
	go test -race -count=1 -coverprofile=coverage.out ./internal/service/...

test-integration:
	go test -race -count=1 -tags=integration ./internal/integration/...

fmt:
	gofmt -w .

lint:
	golangci-lint run ./...

proto-gen:
	buf generate

docker-build:
	docker build -t $(DOCKER_TAG) .

tidy:
	go mod tidy

run: up build
	@echo "Waiting for MySQL to be healthy..."
	@i=0; until docker inspect --format='{{.State.Health.Status}}' leaderboard_mysql 2>/dev/null | grep -q "healthy"; do \
		i=$$((i+1)); \
		if [ $$i -ge 30 ]; then echo " timed out waiting for MySQL"; exit 1; fi; \
		printf "."; sleep 2; \
	done
	@echo " ready."
	@if [ -f .env ]; then export $$(cat .env | grep -v '^#' | xargs) && ./$(BIN); else ./$(BIN); fi

clean:
	rm -f $(BIN) coverage.out
	go clean -testcache -cache
	find . -name '*.test' -type f -delete
