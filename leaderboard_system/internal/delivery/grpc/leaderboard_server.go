package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "leaderboard_system/api/gen/leaderboard/v1"
	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

// mapLeaderboardError maps domain errors to appropriate gRPC status codes.
// Redis connectivity failures return codes.Unavailable so clients know the
// leaderboard is temporarily down while other services (auth, history) remain functional.
func mapLeaderboardError(err error) error {
	if errors.Is(err, domain.ErrRedisUnavailable) {
		return status.Error(codes.Unavailable, "leaderboard service temporarily unavailable")
	}
	return status.Error(codes.Internal, err.Error())
}

// LeaderboardServer implements the LeaderboardService gRPC server.
type LeaderboardServer struct {
	pb.UnimplementedLeaderboardServiceServer
	leaderboard *service.LeaderboardService
}

// NewLeaderboardServer creates a new LeaderboardServer.
func NewLeaderboardServer(lb *service.LeaderboardService) *LeaderboardServer {
	return &LeaderboardServer{leaderboard: lb}
}

func (s *LeaderboardServer) SubmitScore(ctx context.Context, req *pb.SubmitScoreRequest) (*pb.SubmitScoreResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	if err := s.leaderboard.SubmitScore(ctx, userID, req.Game, req.Score); err != nil {
		return nil, mapLeaderboardError(err)
	}

	return &pb.SubmitScoreResponse{Message: "Score submitted successfully"}, nil
}

func (s *LeaderboardServer) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	entries, err := s.leaderboard.GetLeaderboard(ctx, req.Game, req.TopN)
	if err != nil {
		return nil, mapLeaderboardError(err)
	}

	return &pb.GetLeaderboardResponse{Entries: toProtoEntries(entries)}, nil
}

func (s *LeaderboardServer) GetGlobalLeaderboard(ctx context.Context, req *pb.GetGlobalLeaderboardRequest) (*pb.GetGlobalLeaderboardResponse, error) {
	entries, err := s.leaderboard.GetGlobalLeaderboard(ctx, req.TopN)
	if err != nil {
		return nil, mapLeaderboardError(err)
	}

	return &pb.GetGlobalLeaderboardResponse{Entries: toProtoEntries(entries)}, nil
}

func (s *LeaderboardServer) GetUserRank(ctx context.Context, req *pb.GetUserRankRequest) (*pb.GetUserRankResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	entry, err := s.leaderboard.GetUserRank(ctx, req.Game, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user rank not found")
		}
		return nil, mapLeaderboardError(err)
	}

	return &pb.GetUserRankResponse{
		UserId:   entry.UserID,
		Username: entry.Username,
		Score:    entry.Score,
		Rank:     entry.Rank,
	}, nil
}

func (s *LeaderboardServer) GetTopPlayersByPeriod(ctx context.Context, req *pb.GetTopPlayersByPeriodRequest) (*pb.GetTopPlayersByPeriodResponse, error) {
	period := time.Now()
	if req.Period != "" {
		parsed, err := time.Parse("2006-01-02", req.Period)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "period must be in YYYY-MM-DD format")
		}
		period = parsed
	}

	entries, err := s.leaderboard.GetTopPlayersByPeriod(ctx, req.Game, period, req.TopN)
	if err != nil {
		return nil, mapLeaderboardError(err)
	}

	return &pb.GetTopPlayersByPeriodResponse{
		Game:    req.Game,
		Period:  period.Format("2006-01-02"),
		TopN:    req.TopN,
		Entries: toProtoEntries(entries),
	}, nil
}
