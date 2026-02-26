package repository

import (
	"context"
	"database/sql"
	"errors"

	"leaderboard_system/internal/domain"
)

// GameRepo implements domain.GameRepository using MySQL.
type GameRepo struct {
	db *sql.DB
}

// NewGameRepo creates a new GameRepo.
func NewGameRepo(db *sql.DB) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) CreateGame(ctx context.Context, name, description string) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO games (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *GameRepo) ListGames(ctx context.Context) ([]domain.Game, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM games")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (r *GameRepo) GetGameByName(ctx context.Context, name string) (*domain.Game, error) {
	var g domain.Game
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at FROM games WHERE name = ?", name,
	).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGameNotFound
		}
		return nil, err
	}
	return &g, nil
}
