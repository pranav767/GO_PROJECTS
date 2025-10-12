package db

import (
	"leaderboard_system/model"
	"time"
)

// AddScoreHistory stores a score submission in the history table
func AddScoreHistory(userID int64, game string, score float64) error {
	db := GetDB()
	_, err := db.Exec("INSERT INTO score_history (user_id, game, score, submitted_at) VALUES (?, ?, ?, ?)", userID, game, score, time.Now().Format(time.RFC3339))
	return err
}

// GetScoreHistory returns all scores for a user in a game
func GetScoreHistory(userID int64, game string) ([]model.Score, error) {
	db := GetDB()
	rows, err := db.Query("SELECT user_id, game, score, submitted_at FROM score_history WHERE user_id = ? AND game = ? ORDER BY submitted_at DESC", userID, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scores []model.Score
	for rows.Next() {
		var s model.Score
		if err := rows.Scan(&s.UserID, &s.Game, &s.Score, &s.Datetime); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}
