# Example curl Commands

## 1. Register a New User
```
curl -X POST http://localhost:8080/register \
	-H "Content-Type: application/json" \
	-d '{"username": "alice", "password": "password123"}'
```

## 2. Log In (Get JWT Token)
```
curl -X POST http://localhost:8080/login \
	-H "Content-Type: application/json" \
	-d '{"username": "alice", "password": "password123"}'
```
*Save the `token` value from the response for the next steps.*

## 3. Submit a Score (Authenticated)
```
curl -X POST http://localhost:8080/api/submit-score \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer <JWT_TOKEN>" \
	-d '{"game": "chess", "score": 1500}'
```

## 4. Get Leaderboard (Top 10 for a Game)
```
curl -X GET "http://localhost:8080/api/leaderboard?game=chess&topN=10" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 5. Get Your Rank in a Game
```
curl -X GET "http://localhost:8080/api/user-rank?game=chess" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 6. Get Top Players for a Period (Daily)
Default period is today if omitted. Uses max score per user for that day.
```
curl -X GET "http://localhost:8080/api/top-players?game=chess&period=2025-10-04&topN=5" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```
Omitting period (defaults to today):
```
curl -X GET "http://localhost:8080/api/top-players?game=chess&topN=5" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 7. Connect to WebSocket for Real-Time Updates
Use a tool like `wscat`:
```
wscat -c ws://localhost:8080/ws/leaderboard
```

---

Replace `<JWT_TOKEN>` with the token received from the login step.
# Real-Time Leaderboard System (Go + Redis)

> Project idea from: https://roadmap.sh/projects/realtime-leaderboard-system

This project is a backend system for a real-time leaderboard service. It allows users to register, log in, submit scores for games, view leaderboards, and receive real-time updates. Redis sorted sets are used for efficient leaderboard management.


## Features
- **User Authentication**: Register and log in with JWT-based authentication.
- **Game Management**: Create and list games via API.
- **Score Submission**: Submit scores for different games/activities.
- **Score History**: Track and retrieve all score submissions per user/game.
- **Leaderboard**: View global and per-game leaderboards.
- **Global Leaderboard**: See top users across all games.
- **User Ranking**: Query your rank in any game.
- **Top Players Report**: Get top players for a specific period.
- **Real-Time Updates**: WebSocket endpoint for live leaderboard updates.

## Tech Stack & Go Concepts
- Go (Gin, gorilla/websocket, go-redis)
- Redis (via Docker Compose)
- MySQL (for user storage)
- JWT for authentication
- Go routines, channels, mutex for concurrency

---

## Setup Instructions

### 1. Clone the Repository
```bash
git clone <repo-url>
cd leaderboard_system
```

### 2. Start Redis (and MySQL if needed) with Docker Compose
```bash
docker-compose up -d
```

### 3. Configure Environment Variables
Create a `.env` file:
```
HMAC_SECRET=your_jwt_secret
REDIS_ADDR=localhost:6379
```

### 4. Install Go Dependencies
```bash
go mod tidy
```

### 5. Run Database Migrations
Ensure MySQL is running and run the SQL in `internal/db/db.sql` to create the `users` table.

### 6. Start the Server
```bash
go run cmd/main.go
```

---

## API Endpoints


### Auth
- `POST /register` — Register a new user
- `POST /login` — Log in and receive JWT

### Game Management (Protected: JWT required)
- `POST /api/games` — Create a new game `{name, description}`
- `GET /api/games` — List all games

### Leaderboard & Scores (Protected: JWT required)
- `POST /api/submit-score` — Submit a score `{game, score}`
- `GET /api/score-history?user_id=USER_ID&game=GAME` — Get a user's score history for a game
- `GET /api/leaderboard?game=GAME&topN=10` — Get top N leaderboard for a game
- `GET /api/global-leaderboard?topN=10` — Get top N users across all games
- `GET /api/user-rank?game=GAME` — Get your rank in a game
- `GET /api/top-players?game=GAME&period=YYYY-MM-DD&topN=10` — Top players for a period
	*If `period` omitted, defaults to today's date (server time). Aggregation: max score per user for that day.*

### Real-Time Updates
- `GET /ws/leaderboard` — WebSocket endpoint for live leaderboard updates

#### WebSocket Payload Format
Server broadcasts JSON messages shaped like:
```
{
	"type": "leaderboard_update",
	"game": "chess" | "global",
	"entries": [
		{ "user_id": 1, "username": "alice", "score": 1800, "rank": 1 },
		{ "user_id": 2, "username": "bob", "score": 1750, "rank": 2 }
	]
}
```
Clients should ignore unknown fields for forward compatibility.

---


## 1. Register a New User
```
curl -X POST http://localhost:8080/register \
	-H "Content-Type: application/json" \
	-d '{"username": "alice", "password": "password123"}'
```

## 2. Log In (Get JWT Token)
```
curl -X POST http://localhost:8080/login \
	-H "Content-Type: application/json" \
	-d '{"username": "alice", "password": "password123"}'
```
*Save the `token` value from the response for the next steps.*

## 3. Create a Game (Authenticated)
```
curl -X POST http://localhost:8080/api/games \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer <JWT_TOKEN>" \
	-d '{"name": "chess", "description": "Classic board game"}'
```

## 4. List All Games (Authenticated)
```
curl -X GET http://localhost:8080/api/games \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 5. Submit a Score (Authenticated)
```
curl -X POST http://localhost:8080/api/submit-score \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer <JWT_TOKEN>" \
	-d '{"game": "chess", "score": 1500}'
```

## 6. Get Score History for a User/Game (Authenticated)
```
curl -X GET "http://localhost:8080/api/score-history?user_id=1&game=chess" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 7. Get Leaderboard (Top 10 for a Game)
```
curl -X GET "http://localhost:8080/api/leaderboard?game=chess&topN=10" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 8. Get Global Leaderboard (Authenticated)
```
curl -X GET "http://localhost:8080/api/global-leaderboard?topN=10" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 9. Get Your Rank in a Game
```
curl -X GET "http://localhost:8080/api/user-rank?game=chess" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 10. Get Top Players for a Period (Authenticated)
```
curl -X GET "http://localhost:8080/api/top-players?game=chess&period=2025-10-04&topN=5" \
	-H "Authorization: Bearer <JWT_TOKEN>"
```

## 11. Connect to WebSocket for Real-Time Updates
Use a tool like `wscat`:
```
wscat -c ws://localhost:8080/ws/leaderboard
```

---

Replace `<JWT_TOKEN>` with the token received from the login step.
