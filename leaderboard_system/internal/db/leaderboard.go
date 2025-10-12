package db

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)
// GetGlobalLeaderboard returns the top N users in the global leaderboard
func GetGlobalLeaderboard(ctx context.Context, topN int64) ([]redis.Z, error) {
	key := LeaderboardKeyPrefix + "global"
	return RedisClient.ZRevRangeWithScores(ctx, key, 0, topN-1).Result()
}

const (
	LeaderboardKeyPrefix = "leaderboard:"
)

// SubmitScore adds or updates a user's score for a specific game/activity
func SubmitScore(ctx context.Context, game string, userID string, score float64) error {
	key := LeaderboardKeyPrefix + game
	return RedisClient.ZAdd(ctx, key, redis.Z{Score: score, Member: userID}).Err()
}

// GetLeaderboard returns the top N users for a game
func GetLeaderboard(ctx context.Context, game string, topN int64) ([]redis.Z, error) {
	key := LeaderboardKeyPrefix + game
	return RedisClient.ZRevRangeWithScores(ctx, key, 0, topN-1).Result()
}

// GetUserRank returns the rank and score of a user in a game
func GetUserRank(ctx context.Context, game string, userID string) (rank int64, score float64, err error) {
	key := LeaderboardKeyPrefix + game
	rank, err = RedisClient.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return
	}
	score, err = RedisClient.ZScore(ctx, key, userID).Result()
	return
}

// GetTopPlayersByPeriod returns top N users for a game within a time period (assumes time-based keys)
func GetTopPlayersByPeriod(ctx context.Context, game string, period time.Time, topN int64) ([]redis.Z, error) {
	key := fmt.Sprintf("%s%s:%s", LeaderboardKeyPrefix, game, period.Format("2006-01-02"))
	return RedisClient.ZRevRangeWithScores(ctx, key, 0, topN-1).Result()
}
