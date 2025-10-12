
package service

import (
	"leaderboard_system/internal/db"
	"leaderboard_system/model"
)

// CreateGameService creates a new game
func CreateGameService(name, description string) (int64, error) {
	return db.CreateGame(name, description)
}

// ListGamesService lists all games
func ListGamesService() ([]model.Game, error) {
	return db.ListGames()
}
