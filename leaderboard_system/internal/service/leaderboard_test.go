package service_test

import (
	"context"
	"testing"

	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

// mockLeaderboardRepo implements domain.LeaderboardRepository for testing.
type mockLeaderboardRepo struct {
	data map[string]map[string]float64
}

func newMockLeaderboardRepo() *mockLeaderboardRepo {
	return &mockLeaderboardRepo{data: make(map[string]map[string]float64)}
}

func (m *mockLeaderboardRepo) SubmitScore(_ context.Context, key, userID string, score float64) error {
	if m.data[key] == nil {
		m.data[key] = make(map[string]float64)
	}
	m.data[key][userID] = score
	return nil
}

func (m *mockLeaderboardRepo) GetLeaderboard(_ context.Context, key string, topN int64) ([]domain.LeaderboardEntry, error) {
	scores, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	entries := make([]domain.LeaderboardEntry, 0, len(scores))
	rank := int64(1)
	for uid, score := range scores {
		if rank > topN {
			break
		}
		var id int64
		for _, c := range uid {
			id = id*10 + int64(c-'0')
		}
		entries = append(entries, domain.LeaderboardEntry{
			UserID: id,
			Score:  score,
			Rank:   rank,
		})
		rank++
	}
	return entries, nil
}

func (m *mockLeaderboardRepo) GetUserRank(_ context.Context, key, userID string) (int64, float64, error) {
	scores, ok := m.data[key]
	if !ok {
		return -1, 0, domain.ErrUserNotFound
	}
	score, ok := scores[userID]
	if !ok {
		return -1, 0, domain.ErrUserNotFound
	}
	return 1, score, nil
}

// mockGameRepo implements domain.GameRepository for testing.
type mockGameRepo struct {
	games map[string]*domain.Game
}

func newMockGameRepo() *mockGameRepo {
	return &mockGameRepo{games: make(map[string]*domain.Game)}
}

func (m *mockGameRepo) CreateGame(_ context.Context, name, description string) (int64, error) {
	id := int64(len(m.games) + 1)
	m.games[name] = &domain.Game{ID: id, Name: name, Description: description}
	return id, nil
}

func (m *mockGameRepo) ListGames(_ context.Context) ([]domain.Game, error) {
	games := make([]domain.Game, 0, len(m.games))
	for _, g := range m.games {
		games = append(games, *g)
	}
	return games, nil
}

func (m *mockGameRepo) GetGameByName(_ context.Context, name string) (*domain.Game, error) {
	g, ok := m.games[name]
	if !ok {
		return nil, domain.ErrGameNotFound
	}
	return g, nil
}

// mockScoreHistoryRepo implements domain.ScoreHistoryRepository for testing.
type mockScoreHistoryRepo struct {
	history []domain.Score
}

func (m *mockScoreHistoryRepo) AddScoreHistory(_ context.Context, userID int64, game string, score float64) error {
	m.history = append(m.history, domain.Score{UserID: userID, Game: game, Score: score})
	return nil
}

func (m *mockScoreHistoryRepo) GetScoreHistory(_ context.Context, userID int64, game string) ([]domain.Score, error) {
	var result []domain.Score
	for _, s := range m.history {
		if s.UserID == userID && s.Game == game {
			result = append(result, s)
		}
	}
	return result, nil
}

// noopBroadcaster implements service.Broadcaster doing nothing.
type noopBroadcaster struct{}

func (n *noopBroadcaster) Broadcast(_ []byte) {}

func TestSubmitScore_MaxScorePolicy(t *testing.T) {
	lbRepo := newMockLeaderboardRepo()
	gameRepo := newMockGameRepo()
	userRepo := newMockUserRepo()
	historyRepo := &mockScoreHistoryRepo{}

	_, _ = gameRepo.CreateGame(context.Background(), "chess", "board game")
	_, _ = userRepo.CreateUser(context.Background(), "alice", "hash")

	svc := service.NewLeaderboardService(lbRepo, gameRepo, userRepo, historyRepo, &noopBroadcaster{})
	ctx := context.Background()

	// Submit initial score
	err := svc.SubmitScore(ctx, 1, "chess", 1500)
	if err != nil {
		t.Fatalf("submit score failed: %v", err)
	}

	// Submit lower score - should NOT replace
	err = svc.SubmitScore(ctx, 1, "chess", 1000)
	if err != nil {
		t.Fatalf("submit score failed: %v", err)
	}

	// Verify per-game leaderboard still has the higher score
	entry, err := svc.GetUserRank(ctx, "chess", 1)
	if err != nil {
		t.Fatalf("get user rank failed: %v", err)
	}
	if entry.Score != 1500 {
		t.Fatalf("expected score 1500 (max policy), got %f", entry.Score)
	}

	// Submit higher score - should replace
	err = svc.SubmitScore(ctx, 1, "chess", 2000)
	if err != nil {
		t.Fatalf("submit score failed: %v", err)
	}
	entry, err = svc.GetUserRank(ctx, "chess", 1)
	if err != nil {
		t.Fatalf("get user rank failed: %v", err)
	}
	if entry.Score != 2000 {
		t.Fatalf("expected score 2000, got %f", entry.Score)
	}
}

func TestSubmitScore_GameNotFound(t *testing.T) {
	svc := service.NewLeaderboardService(
		newMockLeaderboardRepo(),
		newMockGameRepo(),
		newMockUserRepo(),
		&mockScoreHistoryRepo{},
		&noopBroadcaster{},
	)

	err := svc.SubmitScore(context.Background(), 1, "nonexistent", 100)
	if err == nil {
		t.Fatal("expected error for nonexistent game")
	}
}

func TestGetLeaderboard(t *testing.T) {
	lbRepo := newMockLeaderboardRepo()
	gameRepo := newMockGameRepo()
	userRepo := newMockUserRepo()

	_, _ = gameRepo.CreateGame(context.Background(), "chess", "")
	_, _ = userRepo.CreateUser(context.Background(), "alice", "hash")
	_, _ = userRepo.CreateUser(context.Background(), "bob", "hash")

	svc := service.NewLeaderboardService(lbRepo, gameRepo, userRepo, &mockScoreHistoryRepo{}, &noopBroadcaster{})
	ctx := context.Background()

	_ = svc.SubmitScore(ctx, 1, "chess", 1500)
	_ = svc.SubmitScore(ctx, 2, "chess", 2000)

	entries, err := svc.GetLeaderboard(ctx, "chess", 10)
	if err != nil {
		t.Fatalf("get leaderboard failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetGlobalLeaderboard(t *testing.T) {
	lbRepo := newMockLeaderboardRepo()
	gameRepo := newMockGameRepo()
	userRepo := newMockUserRepo()

	_, _ = gameRepo.CreateGame(context.Background(), "chess", "")
	_, _ = userRepo.CreateUser(context.Background(), "alice", "hash")

	svc := service.NewLeaderboardService(lbRepo, gameRepo, userRepo, &mockScoreHistoryRepo{}, &noopBroadcaster{})
	ctx := context.Background()

	_ = svc.SubmitScore(ctx, 1, "chess", 1500)

	entries, err := svc.GetGlobalLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("get global leaderboard failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}
