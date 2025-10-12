package db

import (
	"leaderboard_system/model"
	"time"
)

// CreateGame inserts a new game into the database
func CreateGame(name, description string) (int64, error) {
	db := GetDB()
	result, err := db.Exec("INSERT INTO games (name, description, created_at) VALUES (?, ?, ?)", name, description, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ListGames returns all games
func ListGames() ([]model.Game, error) {
	db := GetDB()
	rows, err := db.Query("SELECT id, name, description, created_at FROM games")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var games []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}
