package domain

import "context"

// UserRepository defines persistence operations for users.
type UserRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	UpdateUserRole(ctx context.Context, userID int64, role string) error
}

// GameRepository defines persistence operations for games.
type GameRepository interface {
	CreateGame(ctx context.Context, name, description string) (int64, error)
	ListGames(ctx context.Context) ([]Game, error)
	GetGameByName(ctx context.Context, name string) (*Game, error)
}

// LeaderboardRepository defines Redis sorted-set operations for leaderboards.
type LeaderboardRepository interface {
	SubmitScore(ctx context.Context, key, userID string, score float64) error
	GetLeaderboard(ctx context.Context, key string, topN int64) ([]LeaderboardEntry, error)
	GetUserRank(ctx context.Context, key, userID string) (rank int64, score float64, err error)
}

// ScoreHistoryRepository defines persistence operations for score history.
type ScoreHistoryRepository interface {
	AddScoreHistory(ctx context.Context, userID int64, game string, score float64) error
	GetScoreHistory(ctx context.Context, userID int64, game string) ([]Score, error)
}
