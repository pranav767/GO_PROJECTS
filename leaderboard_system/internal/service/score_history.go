package service

import (
	"context"

	"leaderboard_system/internal/domain"
)

// ScoreHistoryService handles score history operations.
type ScoreHistoryService struct {
	history domain.ScoreHistoryRepository
}

// NewScoreHistoryService creates a new ScoreHistoryService.
func NewScoreHistoryService(history domain.ScoreHistoryRepository) *ScoreHistoryService {
	return &ScoreHistoryService{history: history}
}

// GetScoreHistory returns all scores for a user in a game.
func (s *ScoreHistoryService) GetScoreHistory(ctx context.Context, userID int64, game string) ([]domain.Score, error) {
	return s.history.GetScoreHistory(ctx, userID, game)
}
