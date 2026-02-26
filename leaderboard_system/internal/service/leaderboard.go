package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"leaderboard_system/internal/domain"
)

// Broadcaster is implemented by anything that can push messages to connected clients
type Broadcaster interface {
	Broadcast(data []byte)
}

// LeaderboardService handles leaderboard and score submission operations.
type LeaderboardService struct {
	lb      domain.LeaderboardRepository
	games   domain.GameRepository
	users   domain.UserRepository
	history domain.ScoreHistoryRepository
	bc      Broadcaster
}

// NewLeaderboardService creates a new LeaderboardService.
func NewLeaderboardService(
	lb domain.LeaderboardRepository,
	games domain.GameRepository,
	users domain.UserRepository,
	history domain.ScoreHistoryRepository,
	bc Broadcaster,
) *LeaderboardService {
	return &LeaderboardService{lb: lb, games: games, users: users, history: history, bc: bc}
}

// SubmitScore records a score for a user in a game and updates all leaderboards.
func (s *LeaderboardService) SubmitScore(ctx context.Context, userID int64, game string, score float64) error {
	if _, err := s.games.GetGameByName(ctx, game); err != nil {
		return fmt.Errorf("game '%s' does not exist", game)
	}

	userIDStr := int64ToString(userID)

	// Per-game leaderboard: only replace if new score is higher.
	if err := s.submitMaxScore(ctx, game, userIDStr, score); err != nil {
		return err
	}

	// Global leaderboard: only replace if new score beats the global best.
	if err := s.submitMaxScore(ctx, "global", userIDStr, score); err != nil {
		return err
	}

	// Daily leaderboard: max-score per day.
	dayKey := game + ":" + time.Now().Format("2006-01-02")
	if err := s.submitMaxScore(ctx, dayKey, userIDStr, score); err != nil {
		return err
	}

	// Always persist to history (full audit trail).
	if err := s.history.AddScoreHistory(ctx, userID, game, score); err != nil {
		return err
	}

	// Broadcast updated top-10 (best-effort — errors are intentionally ignored).
	s.broadcastTop(ctx, game)
	s.broadcastTop(ctx, "global")

	return nil
}

// GetLeaderboard returns the top N entries for a game leaderboard.
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, game string, topN int64) ([]domain.LeaderboardEntry, error) {
	entries, err := s.lb.GetLeaderboard(ctx, game, topN)
	if err != nil {
		return nil, err
	}
	return s.enrichEntries(ctx, entries), nil
}

// GetGlobalLeaderboard returns the top N entries across all games.
func (s *LeaderboardService) GetGlobalLeaderboard(ctx context.Context, topN int64) ([]domain.LeaderboardEntry, error) {
	entries, err := s.lb.GetLeaderboard(ctx, "global", topN)
	if err != nil {
		return nil, err
	}
	return s.enrichEntries(ctx, entries), nil
}

// GetUserRank returns a user's rank and score for a specific game.
func (s *LeaderboardService) GetUserRank(ctx context.Context, game string, userID int64) (domain.LeaderboardEntry, error) {
	rank, score, err := s.lb.GetUserRank(ctx, game, int64ToString(userID))
	if err != nil {
		return domain.LeaderboardEntry{}, err
	}

	username := ""
	if user, err := s.users.GetUserByID(ctx, userID); err == nil {
		username = user.Username
	}

	return domain.LeaderboardEntry{
		UserID:   userID,
		Username: username,
		Score:    score,
		Rank:     rank + 1,
	}, nil
}

// GetTopPlayersByPeriod returns the top N players for a game on a specific date.
func (s *LeaderboardService) GetTopPlayersByPeriod(ctx context.Context, game string, period time.Time, topN int64) ([]domain.LeaderboardEntry, error) {
	dayKey := game + ":" + period.Format("2006-01-02")
	entries, err := s.lb.GetLeaderboard(ctx, dayKey, topN)
	if err != nil {
		return nil, err
	}
	return s.enrichEntries(ctx, entries), nil
}

// submitMaxScore updates a leaderboard key only when the new score exceeds the existing one.
func (s *LeaderboardService) submitMaxScore(ctx context.Context, key, userIDStr string, score float64) error {
	_, existing, err := s.lb.GetUserRank(ctx, key, userIDStr)
	if err == nil && score <= existing {
		return nil // existing score is higher or equal — keep it
	}
	return s.lb.SubmitScore(ctx, key, userIDStr, score)
}

// enrichEntries fills in Username and Rank fields for a slice of leaderboard entries.
func (s *LeaderboardService) enrichEntries(ctx context.Context, entries []domain.LeaderboardEntry) []domain.LeaderboardEntry {
	for i := range entries {
		entries[i].Rank = int64(i + 1)
		if user, err := s.users.GetUserByID(ctx, entries[i].UserID); err == nil {
			entries[i].Username = user.Username
		}
	}
	return entries
}

// broadcastTop fetches top-10 for a key and broadcasts the JSON payload (best-effort).
func (s *LeaderboardService) broadcastTop(ctx context.Context, key string) {
	entries, err := s.lb.GetLeaderboard(ctx, key, 10)
	if err != nil {
		return
	}
	payload, err := json.Marshal(struct {
		Type    string                    `json:"type"`
		Game    string                    `json:"game"`
		Entries []domain.LeaderboardEntry `json:"entries"`
	}{Type: "leaderboard_update", Game: key, Entries: s.enrichEntries(ctx, entries)})
	if err != nil {
		return
	}
	s.bc.Broadcast(payload)
}

func int64ToString(id int64) string {
	return fmt.Sprintf("%d", id)
}
