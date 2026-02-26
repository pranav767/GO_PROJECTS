package repository

import (
	"context"
	"database/sql"

	"leaderboard_system/internal/domain"
)

// ScoreHistoryRepo implements domain.ScoreHistoryRepository using MySQL.
type ScoreHistoryRepo struct {
	db *sql.DB
}

// NewScoreHistoryRepo creates a new ScoreHistoryRepo.
func NewScoreHistoryRepo(db *sql.DB) *ScoreHistoryRepo {
	return &ScoreHistoryRepo{db: db}
}

func (r *ScoreHistoryRepo) AddScoreHistory(ctx context.Context, userID int64, game string, score float64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO score_history (user_id, game, score) VALUES (?, ?, ?)", userID, game, score)
	return err
}

func (r *ScoreHistoryRepo) GetScoreHistory(ctx context.Context, userID int64, game string) ([]domain.Score, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT user_id, game, score, submitted_at FROM score_history WHERE user_id = ? AND game = ? ORDER BY submitted_at DESC",
		userID, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []domain.Score
	for rows.Next() {
		var s domain.Score
		if err := rows.Scan(&s.UserID, &s.Game, &s.Score, &s.SubmittedAt); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}
