package domain

import "time"

// User represents an authenticated user in the system.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// Game represents a game or activity.
type Game struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// Score represents a single score submission.
type Score struct {
	UserID      int64
	Game        string
	Score       float64
	SubmittedAt time.Time
}

// LeaderboardEntry represents a user's position on a leaderboard.
type LeaderboardEntry struct {
	UserID   int64
	Username string
	Score    float64
	Rank     int64
}
