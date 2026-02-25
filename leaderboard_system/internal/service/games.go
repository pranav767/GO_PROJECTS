package service

import (
	"context"
	"leaderboard_system/internal/db"
	"leaderboard_system/model"
)

// CreateGameService creates a new game
func CreateGameService(ctx context.Context, name, description string) (int64, error) {
	return db.CreateGame(name, description)
}

// ListGamesService lists all games
func ListGamesService(ctx context.Context) ([]model.Game, error) {
	return db.ListGames()
}
