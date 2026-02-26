# Real-Time Leaderboard System

A high-performance leaderboard service built with **gRPC + REST gateway**, **Redis sorted sets**, and **real-time WebSocket** updates.

> Project idea from [roadmap.sh](https://roadmap.sh/projects/realtime-leaderboard-system)

---

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Running the Server](#running-the-server)
- [Trying It Out](#trying-it-out)
- [API Reference](#api-reference)
  - [REST Endpoints](#public-endpoints)
  - [gRPC Services](#grpc-services)
- [WebSocket](#websocket)
- [Health & Observability](#health--observability)
- [Development](#development)
- [Stopping & Cleanup](#stopping--cleanup)
- [Security Notes](#security-notes)
- [Project Structure](#project-structure)

---

## Architecture

```
Client  ──►  gRPC Server   (:9090)  ──►  Services  ──►  Redis  (leaderboards)
        ──►  HTTP Gateway  (:8080)                  ──►  MySQL  (users, games, history)
        ──►  WebSocket     (/ws/leaderboard)
        ──►  Metrics       (/metrics)
        ──►  Health        (/healthz)
```

- **gRPC** is the primary transport; **grpc-gateway** exposes identical REST endpoints on `:8080`
- **Redis sorted sets** power O(log N) leaderboard ops — `ZADD`, `ZREVRANGE`, `ZREVRANK`, `ZSCORE`
- **MySQL** stores users, games, and complete score history (full audit trail)
- **WebSocket hub** broadcasts live top-10 updates on every score submission
- **Protobuf-first** API design with [buf](https://buf.build) + `protovalidate` for input validation at the transport layer
- **Prometheus** metrics collected at both gRPC interceptor and Redis query levels

---

## Features

- **JWT Authentication** — HS256 tokens (24 h expiry), bcrypt password hashing
- **Role-Based Access Control** — `user` and `admin` roles; game creation is admin-only
- **Per-Game Leaderboards** — top N rankings per game, backed by Redis sorted sets
- **Global Leaderboard** — best score per user across all games
- **Daily Leaderboards** — period-keyed sorted sets (`game:YYYY-MM-DD`) for time-range queries
- **Max-Score Policy** — leaderboard only updates when new score exceeds the current best
- **Score History** — every submission logged to MySQL regardless of policy (full audit)
- **Real-Time WebSocket** — top-10 broadcast for affected game + global on every submission
- **Health Checks** — `/healthz` HTTP endpoint (MySQL + Redis ping, returns JSON) + standard `grpc.health.v1` for Kubernetes probes
- **Prometheus Metrics** — gRPC request counts/durations (`grpc_requests_total`, `grpc_request_duration_seconds`) and Redis query durations (`redis_query_duration_seconds`)
- **gRPC Interceptors** — auth, structured logging, Prometheus metrics, panic recovery, proto validation
- **Integration Tests** — real MySQL + Redis via [testcontainers](https://testcontainers.com)
- **Minimal Docker Image** — multi-stage build, scratch runtime, non-root user

---

## Design Decisions

### Why gRPC + grpc-gateway (REST)?

The architecture separates *how services talk to each other* from *how clients talk to the system*.

**gRPC on `:9090` — internal and service-to-service**

- Strongly typed contracts enforced at compile time via generated stubs
- Binary protobuf encoding is significantly faster and smaller than JSON
- If a score submission service, a rewards engine, or a notification service is added later, they call this gRPC server directly — no REST overhead, no serialization round-trip
- Adding a new RPC method to the proto automatically gives any internal caller a typed client for free

**REST on `:8080` (via grpc-gateway) — external client exposure**

- Browsers cannot speak native gRPC (the Fetch API does not support HTTP/2 trailers or binary framing)
- Mobile games and web frontends use whatever HTTP client they already have — no proto toolchain required
- Third-party integrations, webhooks, and admin dashboards all expect JSON over HTTP
- curl, Postman, and any monitoring tool can hit the API without setup

**Why not maintain two separate servers?**

The `.proto` file is the single source of truth. grpc-gateway reads the same proto annotations and generates the REST translation layer — there is no second API definition to keep in sync, no risk of the REST and gRPC contracts drifting apart over time, and no duplicate handler code to maintain. Every new feature added to the proto gets both transports automatically.

> Internal services call `:9090` (gRPC) for performance.
> External clients, browsers, and scripts call `:8080` (REST).

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.25+ |
| Docker & Docker Compose | any recent |
| `make` | any |
| `jq` *(optional)* | pretty-print JSON responses |
| `wscat` *(optional)* | WebSocket testing |

---

## Setup

### 1. Clone

```bash
git clone <repo-url>
cd leaderboard_system
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Required — generate with: openssl rand -hex 32
HMAC_SECRET=change-this-to-a-random-secret

# Redis
REDIS_ADDR=localhost:6379

# MySQL (defaults match docker-compose.yml)
DB_USER=root
DB_PASS=adminpass
DB_HOST=localhost
DB_PORT=3306
DB_NAME=leaderboard_system

# First user registered with this username gets admin role
ADMIN_USERNAME=admin
```

### 3. Start MySQL and Redis

```bash
make up
```

This starts:
- **MySQL 8.0** on `localhost:3306` — schema auto-applied from `internal/db/db.sql`
- **Redis 7.2** on `localhost:6379` — AOF persistence enabled

Wait for MySQL to be healthy (~15 s):

```bash
docker compose ps
# leaderboard_mysql   Up X seconds (healthy)
# leaderboard_redis   Up X seconds
```

---

## Running the Server

```bash
make run
```

Expected output:
```json
{"level":"INFO","msg":"gRPC server starting","port":"9090"}
{"level":"INFO","msg":"HTTP gateway starting","port":"8080"}
{"level":"INFO","msg":"leaderboard system is running","grpc_port":"9090","http_port":"8080"}
```

All REST examples below use the HTTP gateway on `:8080`. The same calls work natively via gRPC on `:9090`.

---

## Trying It Out

### 1. Register admin user

```bash
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq
```
```json
{"message": "User registered successfully"}
```

> If `ADMIN_USERNAME=admin` is set in `.env`, this user is automatically promoted to admin.

### 2. Login and capture token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.token')
```

### 3. Create a game *(admin only)*

```bash
curl -s -X POST http://localhost:8080/api/games \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "chess", "description": "Classic strategy board game"}' | jq
```
```json
{"id": "1", "message": "Game created"}
```

### 4. Register a player and submit scores

```bash
# Register
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "pass1234"}' | jq

# Login
ALICE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "pass1234"}' | jq -r '.token')

# Submit initial score
curl -s -X POST http://localhost:8080/api/submit-score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ALICE" \
  -d '{"game": "chess", "score": 1500}' | jq

# Submit higher score — leaderboard updates
curl -s -X POST http://localhost:8080/api/submit-score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ALICE" \
  -d '{"game": "chess", "score": 1800}' | jq

# Submit lower score — leaderboard does NOT update (max-score policy)
curl -s -X POST http://localhost:8080/api/submit-score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ALICE" \
  -d '{"game": "chess", "score": 1200}' | jq
```

### 5. Query leaderboards

```bash
# Top 10 for chess
curl -s "http://localhost:8080/api/leaderboard?game=chess&top_n=10" \
  -H "Authorization: Bearer $ALICE" | jq

# Global leaderboard (across all games)
curl -s "http://localhost:8080/api/global-leaderboard?top_n=10" \
  -H "Authorization: Bearer $ALICE" | jq

# Your rank in chess
curl -s "http://localhost:8080/api/user-rank?game=chess" \
  -H "Authorization: Bearer $ALICE" | jq

# Daily top players (defaults to today)
curl -s "http://localhost:8080/api/top-players?game=chess&top_n=5" \
  -H "Authorization: Bearer $ALICE" | jq

# Daily top players for a specific date
curl -s "http://localhost:8080/api/top-players?game=chess&period=2026-02-26&top_n=5" \
  -H "Authorization: Bearer $ALICE" | jq

# Score history (all submissions logged, not just leaderboard updates)
curl -s "http://localhost:8080/api/score-history?game=chess" \
  -H "Authorization: Bearer $ALICE" | jq

# Own profile
curl -s "http://localhost:8080/api/profile" \
  -H "Authorization: Bearer $ALICE" | jq
```

### 6. Health check

```bash
curl -s http://localhost:8080/healthz | jq
```
```json
{"mysql": "ok", "redis": "ok"}
```

### 7. Prometheus metrics

```bash
curl -s http://localhost:8080/metrics | grep -E "grpc_requests_total|redis_query"
```

### 8. WebSocket live updates

In a separate terminal, connect **before** submitting a score:

```bash
wscat -c ws://localhost:8080/ws/leaderboard
```

Then submit a score in another terminal — you'll receive:

```json
{
  "type": "leaderboard_update",
  "game": "chess",
  "entries": [{"user_id": 2, "username": "alice", "score": 1800, "rank": 1}]
}
```

Followed immediately by a second broadcast for `"game": "global"`.

---

## API Reference

### Public Endpoints

| Method | Path | Body / Notes | Description |
|--------|------|--------------|-------------|
| `POST` | `/register` | `{"username", "password"}` — min 3 / 6 chars | Create account |
| `POST` | `/login` | `{"username", "password"}` | Returns JWT token |
| `GET` | `/healthz` | — | MySQL + Redis status |
| `GET` | `/ws/leaderboard` | WebSocket upgrade | Live leaderboard feed |
| `GET` | `/metrics` | — | Prometheus metrics |

### Protected Endpoints

All require `Authorization: Bearer <token>` header.

| Method | Path | Role | Query / Body | Description |
|--------|------|------|--------------|-------------|
| `GET` | `/api/profile` | User | — | Own user ID, username, role |
| `POST` | `/api/games` | **Admin** | `{"name", "description"}` | Create a game |
| `GET` | `/api/games` | User | — | List all games |
| `POST` | `/api/submit-score` | User | `{"game", "score"}` | Submit a score |
| `GET` | `/api/leaderboard` | User | `game` (req), `top_n` 1–100 | Top N for a game |
| `GET` | `/api/global-leaderboard` | User | `top_n` 1–100 | Top N across all games |
| `GET` | `/api/user-rank` | User | `game` (req) | Your rank + score |
| `GET` | `/api/top-players` | User | `game` (req), `period` YYYY-MM-DD, `top_n` | Daily top N |
| `GET` | `/api/score-history` | User | `game` (req); `user_id` (admin only) | Full submission history |

### gRPC Services

The server runs on `:9090`. Package: `leaderboard.v1`. Use [grpcurl](https://github.com/fullstorydev/grpcurl) to call methods directly.

**AuthService**

| Method | Full path | Auth | Description |
|--------|-----------|------|-------------|
| `Register` | `/leaderboard.v1.AuthService/Register` | None | Create account |
| `Login` | `/leaderboard.v1.AuthService/Login` | None | Returns JWT token |
| `GetProfile` | `/leaderboard.v1.AuthService/GetProfile` | User | Own profile |

**GameService**

| Method | Full path | Auth | Description |
|--------|-----------|------|-------------|
| `CreateGame` | `/leaderboard.v1.GameService/CreateGame` | Admin | Create a game |
| `ListGames` | `/leaderboard.v1.GameService/ListGames` | User | List all games |

**LeaderboardService**

| Method | Full path | Auth | Description |
|--------|-----------|------|-------------|
| `SubmitScore` | `/leaderboard.v1.LeaderboardService/SubmitScore` | User | Submit a score |
| `GetLeaderboard` | `/leaderboard.v1.LeaderboardService/GetLeaderboard` | User | Top N for a game |
| `GetGlobalLeaderboard` | `/leaderboard.v1.LeaderboardService/GetGlobalLeaderboard` | User | Top N across all games |
| `GetUserRank` | `/leaderboard.v1.LeaderboardService/GetUserRank` | User | Own rank in a game |
| `GetTopPlayersByPeriod` | `/leaderboard.v1.LeaderboardService/GetTopPlayersByPeriod` | User | Daily top N |

**ScoreHistoryService**

| Method | Full path | Auth | Description |
|--------|-----------|------|-------------|
| `GetScoreHistory` | `/leaderboard.v1.ScoreHistoryService/GetScoreHistory` | User | Submission history |

**grpc.health.v1.Health** (standard Kubernetes probe)

| Method | Full path | Auth | Description |
|--------|-----------|------|-------------|
| `Check` | `/grpc.health.v1.Health/Check` | None | Standard gRPC health (updated every 5 s) |

#### grpcurl Examples

```bash
# List all services
grpcurl -plaintext localhost:9090 list

# Register
grpcurl -plaintext -d '{"username":"admin","password":"admin123"}' \
  localhost:9090 leaderboard.v1.AuthService/Register

# Login and capture token
TOKEN=$(grpcurl -plaintext -d '{"username":"admin","password":"admin123"}' \
  localhost:9090 leaderboard.v1.AuthService/Login | jq -r '.token')

# Create a game (admin)
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"name":"chess","description":"Classic board game"}' \
  localhost:9090 leaderboard.v1.GameService/CreateGame

# Submit a score
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"game":"chess","score":1500}' \
  localhost:9090 leaderboard.v1.LeaderboardService/SubmitScore

# Get leaderboard
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"game":"chess","top_n":10}' \
  localhost:9090 leaderboard.v1.LeaderboardService/GetLeaderboard

# Standard gRPC health check
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

### Error Codes

| gRPC Code | HTTP | Meaning |
|-----------|------|---------|
| `UNAUTHENTICATED` | 401 | Missing or invalid JWT |
| `PERMISSION_DENIED` | 403 | Admin-only endpoint |
| `NOT_FOUND` | 404 | Game or user does not exist |
| `INVALID_ARGUMENT` | 400 | Proto validation failure |
| `UNAVAILABLE` | 503 | Redis unreachable |
| `INTERNAL` | 500 | Unexpected server error |

---

## WebSocket

- **Endpoint**: `ws://localhost:8080/ws/leaderboard`
- **Auth**: none required (read-only broadcast)
- **Triggers**: every successful `submit-score` call
- **Broadcasts**: one message for the affected game + one for `"global"`

```json
{
  "type": "leaderboard_update",
  "game": "chess",
  "entries": [
    {"user_id": 1, "username": "alice", "score": 1800, "rank": 1},
    {"user_id": 2, "username": "bob",   "score": 1600, "rank": 2}
  ]
}
```

---

## Health & Observability

### Health Check — `GET /healthz`

Checks MySQL (ping) and Redis (ping). Returns `"ok"` or `"error"` per service.

```json
{"mysql": "ok", "redis": "ok"}
```

Also registered as a standard `grpc.health.v1.Health` service on port `9090` for Kubernetes liveness/readiness probes. The gRPC health status is updated every 5 seconds in the background.

### Prometheus Metrics — `GET /metrics`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `grpc_requests_total` | Counter | `method`, `code` | Total gRPC requests |
| `grpc_request_duration_seconds` | Histogram | `method` | gRPC request latency |
| `redis_query_duration_seconds` | Histogram | `operation` | Redis op latency (`zadd`, `zrevrange`, `zrevrank`) |
| `websocket_active_connections` | Gauge | — | Live WebSocket connections |

---

## Development

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile static binary (`CGO_ENABLED=0`) |
| `make run` | Build and run the server |
| `make up` | Start MySQL + Redis via Docker Compose |
| `make down` | Stop and remove containers (keep volumes) |
| `make reset` | Stop, remove containers and all data volumes |
| `make test` | Unit tests with race detector + coverage |
| `make test-integration` | Integration tests (requires Docker) |
| `make lint` | Run `golangci-lint` |
| `make fmt` | Format all Go files |
| `make vet` | Run `go vet` |
| `make proto-gen` | Regenerate protobuf code via `buf generate` |
| `make docker-build` | Build Docker image (`leaderboard-system:latest`) |
| `make clean` | Remove binary, coverage files, test cache |

### Tests

```bash
# Unit tests — no external dependencies
make test

# View HTML coverage report
go tool cover -html=coverage.out

# Integration tests — spins up real MySQL + Redis via testcontainers
make test-integration
```

### Regenerating Protobuf

Requires [buf CLI](https://buf.build/docs/installation):

```bash
make proto-gen
```

Proto definitions: `api/proto/leaderboard/v1/`
Generated code: `api/gen/leaderboard/v1/` *(do not edit)*

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| API | gRPC, grpc-gateway v2, protobuf (buf) |
| Auth | golang-jwt/v5 (HS256), bcrypt |
| Database | MySQL 8.0 (go-sql-driver) |
| Cache / Leaderboard | Redis 7.2 (go-redis/v9) |
| Real-time | gorilla/websocket |
| Metrics | Prometheus (client_golang) |
| Validation | protovalidate (buf.build) |
| Logging | slog (structured JSON) |
| Testing | testcontainers-go |

---

## Stopping & Cleanup

```bash
# Stop the server
Ctrl+C

# Stop and remove containers (data volumes preserved)
make down

# Full reset — stop, remove containers and all data
make reset

# Start fresh after reset
make up

# Remove build artifacts
make clean
```

---

## Security Notes

- **JWT Secret**: Set a strong `HMAC_SECRET` — generate with `openssl rand -hex 32`
- **Password Hashing**: bcrypt, default cost (10 rounds)
- **SQL Injection**: parameterized queries throughout
- **Input Validation**: `protovalidate` enforces field constraints at the transport layer (min lengths, value ranges)
- **Score Bounds**: `top_n` capped at 100 in proto constraints
- **WebSocket**: public, read-only — no write operations possible
- **Score History**: users can only access their own; admins can query any user via `user_id` param
- **Admin Promotion**: only via `ADMIN_USERNAME` env var — not exposed through any API

---

## Project Structure

```
leaderboard_system/
├── api/
│   ├── proto/leaderboard/v1/        Proto definitions (auth, games, leaderboard, score_history, health)
│   └── gen/leaderboard/v1/          Generated Go code — do not edit
├── cmd/server/main.go               Entry point
├── internal/
│   ├── domain/                      Entities, errors, repository interfaces
│   ├── repository/                  MySQL + Redis implementations
│   ├── service/                     Business logic + unit tests
│   ├── delivery/
│   │   ├── grpc/                    gRPC servers + interceptors (auth, logging, metrics, recovery, validation)
│   │   └── ws/                      WebSocket hub
│   ├── integration/                 End-to-end tests (testcontainers)
│   └── db/db.sql                    MySQL schema
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── .env.example
```

---

## License

MIT

## Acknowledgments

Project idea from [roadmap.sh backend projects](https://roadmap.sh/projects/realtime-leaderboard-system).
