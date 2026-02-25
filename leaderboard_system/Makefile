BIN         := leaderboard-server
DOCKER_TAG  := leaderboard-system:latest

.PHONY: build test test-integration lint proto-gen docker-build run clean fmt vet fuzz help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build             Build static binary (CGO_ENABLED=0)"
	@echo "  run               Build and run the server"
	@echo "  test              Run unit tests with race detector and coverage"
	@echo "  test-integration  Run integration tests (requires Docker)"
	@echo "  fuzz              Fuzz ParseUserID for 30s"
	@echo "  lint              Run golangci-lint"
	@echo "  fmt               Format code with gofmt"
	@echo "  vet               Run go vet"
	@echo "  proto-gen         Regenerate protobuf code (requires buf)"
	@echo "  docker-build      Build Docker image"
	@echo "  clean             Remove build artifacts"

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server/

test:
	go test -race -count=1 -coverprofile=coverage.out ./...

test-integration:
	go test -race -count=1 -tags=integration ./internal/integration/...

lint:
	golangci-lint run ./...

proto-gen:
	buf generate

docker-build:
	docker build -t $(DOCKER_TAG) .

run: build
	./$(BIN)

fmt:
	gofmt -w .

vet:
	go vet ./...

fuzz:
	go test -fuzz=FuzzParseUserID -fuzztime=30s ./internal/service/

clean:
	rm -f $(BIN) coverage.out
