package db

import (
	"leaderboard_system/model"
)

// CreateGame inserts a new game into the database
func CreateGame(name, description string) (int64, error) {
	db := GetDB()
	result, err := db.Exec("INSERT INTO games (name, description) VALUES (?, ?)", name, description)
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

// GetGameByName retrieves a game by its name
func GetGameByName(name string) (*model.Game, error) {
	db := GetDB()
	var g model.Game
	err := db.QueryRow("SELECT id, name, description, created_at FROM games WHERE name = ?", name).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
