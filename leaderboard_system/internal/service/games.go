package service

import (
	"context"
	"log/slog"

	"leaderboard_system/internal/domain"
)

// GameService handles game management operations.
type GameService struct {
	games  domain.GameRepository
	logger *slog.Logger
}

// NewGameService creates a new GameService.
func NewGameService(games domain.GameRepository, logger *slog.Logger) *GameService {
	return &GameService{games: games, logger: logger}
}

// CreateGame creates a new game.
func (s *GameService) CreateGame(ctx context.Context, name, description string) (int64, error) {
	s.logger.Info("creating game", slog.String("name", name))
	return s.games.CreateGame(ctx, name, description)
}

// ListGames returns all games.
func (s *GameService) ListGames(ctx context.Context) ([]domain.Game, error) {
	return s.games.ListGames(ctx)
}
