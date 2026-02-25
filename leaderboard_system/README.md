# Real-Time Leaderboard System

> A backend system for managing real-time competitive leaderboards with user authentication, score tracking, and live WebSocket updates.
>
> Project idea from [roadmap.sh](https://roadmap.sh/projects/realtime-leaderboard-system)

## Overview

This project implements a high-performance real-time leaderboard service where users can compete in various games or activities, submit scores, and view their rankings. The system features user authentication with role-based access control, score submission, real-time leaderboard updates via WebSocket, and comprehensive score history tracking.

Redis sorted sets power the leaderboard functionality, providing efficient O(log N) operations for score updates and rank queries, while MySQL handles persistent storage of users, games, and score history.

## Features

- **User Authentication**: Secure registration and JWT-based login with role support (user/admin)
- **Role-Based Access Control**: Admin-only game creation, protected API endpoints
- **Game Management**: Create and list games/activities (admin only for creation)
- **Score Submission**: Submit scores for games with automatic max-score aggregation
- **Score History**: Track complete submission history per user and game
- **Per-Game Leaderboards**: Top N rankings for individual games
- **Global Leaderboard**: Aggregate top performers across all games
- **User Rankings**: Query personal rank and score in any game
- **Period-Based Reports**: Daily top players using time-keyed sorted sets
- **Real-Time Updates**: WebSocket broadcasts for live leaderboard changes
- **Input Validation**: Protected against DoS via topN limits and game validation

## Tech Stack

- **Go 1.24+** with Gin web framework
- **MySQL 8.0** for persistent user, game, and history storage
- **Redis 7.2** with sorted sets for leaderboard operations
- **JWT (golang-jwt/v5)** for stateless authentication
- **WebSocket (gorilla/websocket)** for real-time updates
- **Docker Compose** for local development environment

## Architecture

- **HTTP REST API**: Gin router with JWT middleware for authentication
- **Redis Sorted Sets**: Keys like `leaderboard:chess`, `leaderboard:global`, `leaderboard:chess:2026-02-25`
  - Members: user IDs (numeric strings)
  - Scores: floating point values
  - Operations: `ZADD`, `ZREVRANGE`, `ZREVRANK`, `ZSCORE`
- **MySQL Tables**: `users` (auth + roles), `games` (metadata), `score_history` (audit log)
- **Concurrency**: Goroutines for WebSocket broadcaster, mutex-protected client registry
- **Score Policy**: Max score per user per game (only higher scores update leaderboards)

## Prerequisites

- **Go 1.24+** installed
- **Docker & Docker Compose** for MySQL and Redis
- **wscat** or similar for WebSocket testing (optional)

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repo-url>
cd leaderboard_system
```

### 2. Start MySQL and Redis

```bash
docker compose up -d
```

This starts:
- MySQL 8.0 on port 3306 (auto-runs `internal/db/db.sql` schema)
- Redis 7.2 on port 6379 with AOF persistence

### 3. Configure Environment Variables

Create a `.env` file in the project root:

```env
# JWT Secret (required - use a strong random string in production)
HMAC_SECRET=your_secret_key_change_this_in_production

# Redis Configuration
REDIS_ADDR=localhost:6379

# MySQL Configuration (optional - defaults shown)
DB_USER=root
DB_PASS=adminpass
DB_HOST=localhost
DB_PORT=3306
DB_NAME=leaderboard_system

# Admin User (optional - set to auto-promote first registered user with this username)
ADMIN_USERNAME=admin
```

**Note**: If `ADMIN_USERNAME` is set, the first user registered with that username will be automatically promoted to admin role.

### 4. Install Go Dependencies

```bash
go mod download
```

### 5. Run the Server

```bash
go run cmd/main.go
```

Server starts on `http://localhost:8080`

### 6. Create the First Admin User

```bash
# Register with the username matching ADMIN_USERNAME
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "securepass123"}'
```

## API Reference

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/register` | Register a new user |
| `POST` | `/login` | Authenticate and receive JWT token |
| `GET` | `/ws/leaderboard` | WebSocket for real-time leaderboard updates |

### Protected Endpoints (JWT Required)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/games` | **Admin** | Create a new game |
| `GET` | `/api/games` | User | List all games |
| `POST` | `/api/submit-score` | User | Submit a score for a game |
| `GET` | `/api/score-history` | User | Get own score history (admin can query any user) |
| `GET` | `/api/leaderboard` | User | Get top N for a game |
| `GET` | `/api/global-leaderboard` | User | Get top N across all games |
| `GET` | `/api/user-rank` | User | Get your rank in a game |
| `GET` | `/api/top-players` | User | Get top N for a game within a time period |

### Endpoint Details

#### `POST /register`
Register a new user.

**Request Body:**
```json
{
  "username": "alice",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "User registered successfully"
}
```

---

#### `POST /login`
Authenticate and receive a JWT token.

**Request Body:**
```json
{
  "username": "alice",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Use this token in the `Authorization: Bearer <token>` header for protected endpoints.

---

#### `POST /api/games` (Admin Only)
Create a new game.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Request Body:**
```json
{
  "name": "chess",
  "description": "Classic strategy board game"
}
```

**Response:**
```json
{
  "id": 1,
  "message": "Game created"
}
```

---

#### `GET /api/games`
List all available games.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Response:**
```json
[
  {
    "id": 1,
    "name": "chess",
    "description": "Classic strategy board game",
    "created_at": "2026-02-25T10:30:00Z"
  }
]
```

---

#### `POST /api/submit-score`
Submit a score for a game. Only scores higher than your current best are recorded on the leaderboard (max score policy).

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Request Body:**
```json
{
  "game": "chess",
  "score": 1500
}
```

**Response:**
```json
{
  "message": "Score submitted"
}
```

**Notes:**
- Game must exist in the database (created via `/api/games`)
- All submissions are logged in score history
- Leaderboards only update if the new score is higher than the previous best

---

#### `GET /api/score-history?game=<GAME>`
Get your score history for a game. Admins can query any user with `?user_id=<ID>`.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `game` (required): Game name
- `user_id` (admin only): Query another user's history

**Response:**
```json
[
  {
    "user_id": 1,
    "game": "chess",
    "score": 1500,
    "datetime": "2026-02-25T10:45:00Z"
  },
  {
    "user_id": 1,
    "game": "chess",
    "score": 1450,
    "datetime": "2026-02-24T14:20:00Z"
  }
]
```

---

#### `GET /api/leaderboard?game=<GAME>&topN=<N>`
Get the top N players for a specific game.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `game` (required): Game name
- `topN` (optional): Number of results (default: 10, max: 100)

**Response:**
```json
[
  {
    "user_id": 5,
    "username": "alice",
    "score": 1800,
    "rank": 1
  },
  {
    "user_id": 3,
    "username": "bob",
    "score": 1750,
    "rank": 2
  }
]
```

---

#### `GET /api/global-leaderboard?topN=<N>`
Get the top N players across all games (based on each user's highest score in any game).

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `topN` (optional): Number of results (default: 10, max: 100)

**Response:**
```json
[
  {
    "user_id": 5,
    "username": "alice",
    "score": 1800,
    "rank": 1
  }
]
```

---

#### `GET /api/user-rank?game=<GAME>`
Get your rank and score in a specific game.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `game` (required): Game name

**Response:**
```json
{
  "user_id": 5,
  "username": "alice",
  "score": 1500,
  "rank": 3
}
```

---

#### `GET /api/top-players?game=<GAME>&period=<YYYY-MM-DD>&topN=<N>`
Get the top N players for a specific game within a time period (daily).

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Query Parameters:**
- `game` (required): Game name
- `period` (optional): Date in YYYY-MM-DD format (defaults to today)
- `topN` (optional): Number of results (default: 10, max: 100)

**Response:**
```json
{
  "game": "chess",
  "period": "2026-02-25",
  "topN": 5,
  "entries": [
    {
      "user_id": 5,
      "username": "alice",
      "score": 1500,
      "rank": 1
    }
  ]
}
```

---

### WebSocket Endpoint

#### `GET /ws/leaderboard`
Connect to receive real-time leaderboard updates via WebSocket.

**Example using wscat:**
```bash
wscat -c ws://localhost:8080/ws/leaderboard
```

**Broadcast Message Format:**
```json
{
  "type": "leaderboard_update",
  "game": "chess",
  "entries": [
    {
      "user_id": 5,
      "username": "alice",
      "score": 1800,
      "rank": 1
    },
    {
      "user_id": 3,
      "username": "bob",
      "score": 1750,
      "rank": 2
    }
  ]
}
```

**Notes:**
- No authentication required (read-only broadcast)
- Server broadcasts top 10 for affected game + global leaderboard on every score submission
- `game` field is either a game name or `"global"`

## Usage Examples

### Complete Workflow

```bash
# 1. Register as admin
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 2. Login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.token')

# 3. Create a game (admin only)
curl -X POST http://localhost:8080/api/games \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "chess", "description": "Classic board game"}'

# 4. Register a regular user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "pass123"}'

# 5. Login as alice
ALICE_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "pass123"}' | jq -r '.token')

# 6. Submit a score
curl -X POST http://localhost:8080/api/submit-score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -d '{"game": "chess", "score": 1500}'

# 7. Get leaderboard
curl -X GET "http://localhost:8080/api/leaderboard?game=chess&topN=10" \
  -H "Authorization: Bearer $ALICE_TOKEN"

# 8. Get your rank
curl -X GET "http://localhost:8080/api/user-rank?game=chess" \
  -H "Authorization: Bearer $ALICE_TOKEN"

# 9. Get global leaderboard
curl -X GET "http://localhost:8080/api/global-leaderboard?topN=10" \
  -H "Authorization: Bearer $ALICE_TOKEN"

# 10. View your score history
curl -X GET "http://localhost:8080/api/score-history?game=chess" \
  -H "Authorization: Bearer $ALICE_TOKEN"

# 11. Get top players for today
curl -X GET "http://localhost:8080/api/top-players?game=chess&topN=5" \
  -H "Authorization: Bearer $ALICE_TOKEN"

# 12. Connect to WebSocket for real-time updates
wscat -c ws://localhost:8080/ws/leaderboard
```

## Project Requirements Checklist

✅ **User Authentication**: JWT-based register + login with bcrypt password hashing  
✅ **Score Submission**: Users submit scores tied to their authenticated identity  
✅ **Leaderboard Updates**: Global leaderboard shows top users across all games  
✅ **User Rankings**: Users can query their rank in any game  
✅ **Top Players Report**: Generate daily reports using period-keyed sorted sets  
✅ **Redis Sorted Sets**: All leaderboard data stored and queried via Redis sorted sets  
✅ **Real-Time Updates**: WebSocket broadcasts leaderboard changes on every score submission  
✅ **Efficient Rank Queries**: O(log N) operations using `ZREVRANK`, `ZREVRANGE`, `ZSCORE`

## Additional Features Beyond Requirements

- **Role-Based Access Control**: Admin role with restricted game creation
- **Score History Audit**: Complete submission log in MySQL
- **Game Validation**: Scores only accepted for existing games
- **Input Sanitization**: TopN capped at 100, defaults to 10 on invalid input
- **Environment-Based Configuration**: All credentials and settings via .env
- **Max Score Aggregation**: Automatic best-score policy per user per game
- **Period Leaderboards**: Daily time-series sorted sets for historical reports

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o leaderboard-server cmd/main.go
./leaderboard-server
```

### Docker Compose Services
```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down

# Reset database
docker compose down -v
docker compose up -d
```

## Security Notes

- **JWT Secret**: Set a strong `HMAC_SECRET` in production (use `openssl rand -hex 32`)
- **Password Hashing**: bcrypt with default cost (10 rounds)
- **SQL Injection**: Parameterized queries throughout
- **WebSocket**: Public endpoint (no auth) - suitable for read-only leaderboard viewing
- **Score History**: Access restricted to own user unless admin
- **Admin Promotion**: Only via environment variable, not via API

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Acknowledgments

Project idea from [roadmap.sh backend projects](https://roadmap.sh/projects/realtime-leaderboard-system)
