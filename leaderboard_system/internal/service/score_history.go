
package service

import (
	"leaderboard_system/internal/db"
	"leaderboard_system/model"
)

// AddScoreHistoryService stores a score submission in the history table
func AddScoreHistoryService(userID int64, game string, score float64) error {
	return db.AddScoreHistory(userID, game, score)
}

// GetScoreHistoryService returns all scores for a user in a game
func GetScoreHistoryService(userID int64, game string) ([]model.Score, error) {
	return db.GetScoreHistory(userID, game)
}
