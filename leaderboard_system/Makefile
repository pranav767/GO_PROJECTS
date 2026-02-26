BIN         := leaderboard-server
DOCKER_TAG  := leaderboard-system:latest

.PHONY: build test test-integration lint install-lint proto-gen docker-build run clean fmt vet up down reset help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build             Build static binary (CGO_ENABLED=0)"
	@echo "  run               Build and run the server"
	@echo "  up                Start MySQL and Redis via Docker Compose"
	@echo "  down              Stop Docker Compose services (keep data)"
	@echo "  reset             Stop services and remove all volumes (full reset)"
	@echo "  test              Run unit tests with race detector and coverage"
	@echo "  test-integration  Run integration tests (requires Docker)"
	@echo "  lint              Run golangci-lint"
	@echo "  install-lint      Install golangci-lint"
	@echo "  fmt               Format code with gofmt"
	@echo "  vet               Run go vet"
	@echo "  proto-gen         Regenerate protobuf code (requires buf)"
	@echo "  docker-build      Build Docker image"
	@echo "  clean             Remove binaries, test cache, and coverage files"

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
	go test -race -count=1 -coverprofile=coverage.out \
		./internal/service/... \
		./internal/delivery/grpc/interceptors/...

test-integration:
	go test -race -count=1 -tags=integration ./internal/integration/...

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run ./...

proto-gen:
	buf generate

docker-build:
	docker build -t $(DOCKER_TAG) .

run: build
	@echo "Waiting for MySQL to be healthy..."
	@until docker inspect --format='{{.State.Health.Status}}' leaderboard_mysql 2>/dev/null | grep -q "healthy"; do \
		printf "."; sleep 2; \
	done
	@echo " ready."
	@if [ -f .env ]; then export $$(cat .env | grep -v '^#' | xargs) && ./$(BIN); else ./$(BIN); fi

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -f $(BIN) coverage.out
	go clean -testcache -cache
	find . -name '*.test' -type f -delete
	find . -name '*.out' -type f -delete
