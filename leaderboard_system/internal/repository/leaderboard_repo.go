package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	"leaderboard_system/internal/domain"
)

// isRedisUnavailable checks whether an error indicates Redis is unreachable.
func isRedisUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, redis.ErrClosed) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

// wrapRedisError converts Redis connectivity errors into domain.ErrRedisUnavailable.
func wrapRedisError(err error) error {
	if isRedisUnavailable(err) {
		return fmt.Errorf("%w: %v", domain.ErrRedisUnavailable, err)
	}
	return err
}

const leaderboardKeyPrefix = "leaderboard:"

// LeaderboardRepo implements domain.LeaderboardRepository using Redis sorted sets.
type LeaderboardRepo struct {
	client *redis.Client
}

// NewLeaderboardRepo creates a new LeaderboardRepo.
func NewLeaderboardRepo(client *redis.Client) *LeaderboardRepo {
	return &LeaderboardRepo{client: client}
}

func (r *LeaderboardRepo) SubmitScore(ctx context.Context, key, userID string, score float64) error {
	start := time.Now()
	defer func() {
		redisQueryDuration.WithLabelValues("zadd").Observe(time.Since(start).Seconds())
	}()

	fullKey := leaderboardKeyPrefix + key
	return wrapRedisError(r.client.ZAdd(ctx, fullKey, redis.Z{Score: score, Member: userID}).Err())
}

func (r *LeaderboardRepo) GetLeaderboard(ctx context.Context, key string, topN int64) ([]domain.LeaderboardEntry, error) {
	start := time.Now()
	defer func() {
		redisQueryDuration.WithLabelValues("zrevrange").Observe(time.Since(start).Seconds())
	}()

	fullKey := leaderboardKeyPrefix + key
	results, err := r.client.ZRevRangeWithScores(ctx, fullKey, 0, topN-1).Result()
	if err != nil {
		return nil, wrapRedisError(err)
	}

	entries := make([]domain.LeaderboardEntry, 0, len(results))
	for i, z := range results {
		uid, err := strconv.ParseInt(fmt.Sprint(z.Member), 10, 64)
		if err != nil {
			uid = 0
		}
		entries = append(entries, domain.LeaderboardEntry{
			UserID: uid,
			Score:  z.Score,
			Rank:   int64(i + 1),
		})
	}
	return entries, nil
}

func (r *LeaderboardRepo) GetUserRank(ctx context.Context, key, userID string) (int64, float64, error) {
	start := time.Now()
	defer func() {
		redisQueryDuration.WithLabelValues("zrevrank").Observe(time.Since(start).Seconds())
	}()

	fullKey := leaderboardKeyPrefix + key
	rank, err := r.client.ZRevRank(ctx, fullKey, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, 0, domain.ErrUserNotFound
		}
		return -1, 0, wrapRedisError(err)
	}
	score, err := r.client.ZScore(ctx, fullKey, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, 0, domain.ErrUserNotFound
		}
		return -1, 0, wrapRedisError(err)
	}
	return rank + 1, score, nil // +1 to convert 0-based Redis rank to 1-based display rank
}

var redisQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "redis_query_duration_seconds",
		Help:    "Redis query duration in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"operation"},
)

func init() {
	prometheus.MustRegister(redisQueryDuration)
}
