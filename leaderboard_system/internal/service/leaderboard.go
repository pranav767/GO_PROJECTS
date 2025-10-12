package service

import (
	"context"
	"encoding/json"
	"fmt"
	"leaderboard_system/internal/db"
	"leaderboard_system/internal/realtime"
	"leaderboard_system/model"
	"time"
)

// GetGlobalLeaderboardService returns the top N global leaderboard entries
func GetGlobalLeaderboardService(ctx context.Context, topN int64) ([]model.LeaderboardEntry, error) {
	entries, err := db.GetGlobalLeaderboard(ctx, topN)
	if err != nil {
		return nil, err
	}
	var result []model.LeaderboardEntry
	for i, z := range entries {
		// Debug: print z.Member value and type
		//fmt.Printf("[DEBUG] z.Member: value=%v, type=%T\n", z.Member, z.Member)
		userID, _ := parseUserID(z.Member)
		//fmt.Printf("[DEBUG] parsed userID: %d\n", userID)
		username := ""
		if user, err := db.GetUserByID(userID); err == nil {
			username = user.Username
		}
		result = append(result, model.LeaderboardEntry{
			UserID:   userID,
			Username: username,
			Score:    z.Score,
			Rank:     int64(i + 1),
		})
	}
	return result, nil
}

// SubmitScoreService handles score submission and leaderboard update
func SubmitScoreService(ctx context.Context, userID int64, game string, score float64) error {
	userIDStr := int64ToString(userID)
	// 1. Update per-game leaderboard (use max score aggregation)
	// Fetch existing score (if any) and only replace if incoming is higher
	if existingRank, existingScore, err := db.GetUserRank(ctx, game, userIDStr); err == nil && existingRank >= 0 {
		if score < existingScore {
			// Keep higher (max) score policy
		} else {
			if err := db.SubmitScore(ctx, game, userIDStr, score); err != nil {
				return err
			}
		}
	} else {
		if err := db.SubmitScore(ctx, game, userIDStr, score); err != nil {
			return err
		}
	}

	// 2. Update global leaderboard (max across all games)
	if existingRank, existingScore, err := db.GetUserRank(ctx, "global", userIDStr); err == nil && existingRank >= 0 {
		if score > existingScore { // replace only if better than previous global best
			if err := db.SubmitScore(ctx, "global", userIDStr, score); err != nil {
				return err
			}
		}
	} else {
		if err := db.SubmitScore(ctx, "global", userIDStr, score); err != nil {
			return err
		}
	}

	// 3. Update daily (period) leaderboard (date key) always with max semantics per day
	dayKey := game + ":" + time.Now().Format("2006-01-02")
	if existingRank, existingScore, err := db.GetUserRank(ctx, dayKey, userIDStr); err == nil && existingRank >= 0 {
		if score > existingScore {
			if err := db.SubmitScore(ctx, dayKey, userIDStr, score); err != nil {
				return err
			}
		}
	} else {
		if err := db.SubmitScore(ctx, dayKey, userIDStr, score); err != nil {
			return err
		}
	}

	// 4. Persist score history (store every submission, even if it doesn't change leaderboard)
	if err := db.AddScoreHistory(userID, game, score); err != nil {
		return err
	}

	// 5. Broadcast updated top N for the game and global (best effort)
	broadcastTop := int64(10)
	if gameEntries, err := GetLeaderboardService(ctx, game, broadcastTop); err == nil {
		if payload, err := json.Marshal(struct {
			Type    string                   `json:"type"`
			Game    string                   `json:"game"`
			Entries []model.LeaderboardEntry `json:"entries"`
		}{Type: "leaderboard_update", Game: game, Entries: gameEntries}); err == nil {
			realtime.Broadcast(payload)
		}
	}
	if globalEntries, err := GetGlobalLeaderboardService(ctx, broadcastTop); err == nil {
		if payload, err := json.Marshal(struct {
			Type    string                   `json:"type"`
			Game    string                   `json:"game"`
			Entries []model.LeaderboardEntry `json:"entries"`
		}{Type: "leaderboard_update", Game: "global", Entries: globalEntries}); err == nil {
			realtime.Broadcast(payload)
		}
	}
	return nil
}

// GetLeaderboardService returns the top N leaderboard entries for a game
func GetLeaderboardService(ctx context.Context, game string, topN int64) ([]model.LeaderboardEntry, error) {
	entries, err := db.GetLeaderboard(ctx, game, topN)
	if err != nil {
		return nil, err
	}
	var result []model.LeaderboardEntry
	for i, z := range entries {
		// Debug: print z.Member value and type
		//fmt.Printf("[DEBUG] z.Member: value=%v, type=%T\n", z.Member, z.Member)
		userID, _ := parseUserID(z.Member)
		//fmt.Printf("[DEBUG] parsed userID: %d\n", userID)
		username := ""
		if user, err := db.GetUserByID(userID); err == nil {
			username = user.Username
		}
		result = append(result, model.LeaderboardEntry{
			UserID:   userID,
			Username: username,
			Score:    z.Score,
			Rank:     int64(i + 1),
		})
	}
	return result, nil
}

// GetUserRankService returns a user's rank and score for a game
func GetUserRankService(ctx context.Context, game string, userID int64) (model.LeaderboardEntry, error) {
	rank, score, err := db.GetUserRank(ctx, game, int64ToString(userID))
	if err != nil {
		return model.LeaderboardEntry{}, err
	}
	username := ""
	if user, err := db.GetUserByID(userID); err == nil {
		username = user.Username
	}
	return model.LeaderboardEntry{UserID: userID, Username: username, Score: score, Rank: rank + 1}, nil
}

// GetTopPlayersByPeriodService returns top N players for a game in a period
func GetTopPlayersByPeriodService(ctx context.Context, game string, period time.Time, topN int64) ([]model.LeaderboardEntry, error) {
	entries, err := db.GetTopPlayersByPeriod(ctx, game, period, topN)
	if err != nil {
		return nil, err
	}
	var result []model.LeaderboardEntry
	for i, z := range entries {
		// Debug: print z.Member value and type
		// fmt.Printf("[DEBUG] z.Member: value=%v, type=%T\n", z.Member, z.Member)
		userID, _ := parseUserID(z.Member)
		// fmt.Printf("[DEBUG] parsed userID: %d\n", userID)
		username := ""
		if user, err := db.GetUserByID(userID); err == nil {
			username = user.Username
		}
		result = append(result, model.LeaderboardEntry{
			UserID:   userID,
			Username: username,
			Score:    z.Score,
			Rank:     int64(i + 1),
		})
	}
	return result, nil
}

// parseUserID tries to parse a Redis member (string) to int64 userID
func parseUserID(member interface{}) (int64, error) {
	switch v := member.(type) {
	case string:
		return stringToInt64(v)
	case []byte:
		return stringToInt64(string(v))
	default:
		return 0, nil
	}
}

func stringToInt64(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscan(s, &id)
	return id, err
}

func int64ToString(id int64) string {
	return fmt.Sprintf("%d", id)
}
