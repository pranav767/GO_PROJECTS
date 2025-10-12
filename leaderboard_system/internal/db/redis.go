package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

var RedisClient *redis.Client

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     getRedisAddr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	return err
}

func getRedisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}
