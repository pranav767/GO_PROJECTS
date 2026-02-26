package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "leaderboard_system/api/gen/leaderboard/v1"
	"leaderboard_system/internal/service"
)

// ScoreHistoryServer implements the ScoreHistoryService gRPC server.
type ScoreHistoryServer struct {
	pb.UnimplementedScoreHistoryServiceServer
	history *service.ScoreHistoryService
}

// NewScoreHistoryServer creates a new ScoreHistoryServer.
func NewScoreHistoryServer(history *service.ScoreHistoryService) *ScoreHistoryServer {
	return &ScoreHistoryServer{history: history}
}

func (s *ScoreHistoryServer) GetScoreHistory(ctx context.Context, req *pb.GetScoreHistoryRequest) (*pb.GetScoreHistoryResponse, error) {
	authenticatedUserID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	// Determine which user's history to fetch
	targetUserID := authenticatedUserID
	if req.UserId != 0 && req.UserId != authenticatedUserID {
		// Only admins can query other users' history
		if roleFromContext(ctx) != "admin" {
			return nil, status.Error(codes.PermissionDenied, "admin access required to view other users' history")
		}
		targetUserID = req.UserId
	}

	scores, err := s.history.GetScoreHistory(ctx, targetUserID, req.Game)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	entries := make([]*pb.ScoreEntry, 0, len(scores))
	for _, score := range scores {
		entries = append(entries, &pb.ScoreEntry{
			UserId:   score.UserID,
			Game:     score.Game,
			Score:    score.Score,
			Datetime: score.SubmittedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.GetScoreHistoryResponse{Scores: entries}, nil
}
