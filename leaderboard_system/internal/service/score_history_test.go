package service_test

import (
	"context"
	"testing"

	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

func TestGetScoreHistory(t *testing.T) {
	repo := &mockScoreHistoryRepo{}
	svc := service.NewScoreHistoryService(repo)
	ctx := context.Background()

	// Seed history directly via the repository (AddScoreHistory is a repo concern)
	repo.history = append(repo.history,
		domain.Score{UserID: 1, Game: "chess", Score: 1500},
		domain.Score{UserID: 1, Game: "chess", Score: 2000},
	)

	scores, err := svc.GetScoreHistory(ctx, 1, "chess")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
}

func TestGetScoreHistory_NoResults(t *testing.T) {
	repo := &mockScoreHistoryRepo{}
	svc := service.NewScoreHistoryService(repo)

	scores, err := svc.GetScoreHistory(context.Background(), 99, "nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(scores) != 0 {
		t.Fatalf("expected 0 scores, got %d", len(scores))
	}
}
