package service

import (
	"context"
	"leaderboard_system/internal/db"
	"leaderboard_system/model"
)

// AddScoreHistoryService stores a score submission in the history table
func AddScoreHistoryService(ctx context.Context, userID int64, game string, score float64) error {
	return db.AddScoreHistory(ctx, userID, game, score)
}

// GetScoreHistoryService returns all scores for a user in a game
func GetScoreHistoryService(ctx context.Context, userID int64, game string) ([]model.Score, error) {
	return db.GetScoreHistory(ctx, userID, game)
}
