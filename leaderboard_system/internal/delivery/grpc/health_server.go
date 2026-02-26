package grpc

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "leaderboard_system/api/gen/leaderboard/v1"
)

// HealthServer implements the HealthService gRPC server.
type HealthServer struct {
	pb.UnimplementedHealthServiceServer
	db    *sql.DB
	redis *redis.Client
}

// NewHealthServer creates a new HealthServer.
func NewHealthServer(db *sql.DB, redis *redis.Client) *HealthServer {
	return &HealthServer{db: db, redis: redis}
}

func (s *HealthServer) Check(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	resp := &pb.HealthCheckResponse{
		Mysql: "healthy",
		Redis: "healthy",
	}

	if err := s.db.PingContext(ctx); err != nil {
		resp.Mysql = "unhealthy: " + err.Error()
	}

	if _, err := s.redis.Ping(ctx).Result(); err != nil {
		resp.Redis = "unhealthy: " + err.Error()
	}

	if resp.Mysql != "healthy" || resp.Redis != "healthy" {
		return resp, status.Error(codes.Unavailable, "one or more services are unhealthy")
	}

	return resp, nil
}
