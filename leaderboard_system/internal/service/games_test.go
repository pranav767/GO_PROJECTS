package service_test

import (
	"context"
	"testing"

	"leaderboard_system/internal/service"
)

func TestCreateGame(t *testing.T) {
	repo := newMockGameRepo()
	svc := service.NewGameService(repo, testLogger())

	id, err := svc.CreateGame(context.Background(), "chess", "a board game")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestListGames(t *testing.T) {
	repo := newMockGameRepo()
	svc := service.NewGameService(repo, testLogger())

	_, _ = svc.CreateGame(context.Background(), "chess", "board game")
	_, _ = svc.CreateGame(context.Background(), "tetris", "puzzle game")

	games, err := svc.ListGames(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}
}
