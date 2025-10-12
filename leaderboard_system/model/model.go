package model

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Score represents a user's score submission for a game/activity
type Score struct {
	UserID   int64   `json:"user_id"`
	Game     string  `json:"game"`
	Score    float64 `json:"score"`
	Datetime string  `json:"datetime"` // ISO8601 format
}

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
	Rank     int64   `json:"rank"`
}

// ScoreHistory represents a user's score history for a game
type ScoreHistory struct {
	UserID   int64   `json:"user_id"`
	Game     string  `json:"game"`
	Scores   []Score `json:"scores"`
}

// Game represents a game or activity in the system
type Game struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}