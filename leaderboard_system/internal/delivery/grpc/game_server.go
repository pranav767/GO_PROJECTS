package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "leaderboard_system/api/gen/leaderboard/v1"
	"leaderboard_system/internal/service"
)

// GameServer implements the GameService gRPC server.
type GameServer struct {
	pb.UnimplementedGameServiceServer
	games *service.GameService
}

// NewGameServer creates a new GameServer.
func NewGameServer(games *service.GameService) *GameServer {
	return &GameServer{games: games}
}

func (s *GameServer) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	if roleFromContext(ctx) != "admin" {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	id, err := s.games.CreateGame(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateGameResponse{Id: id, Message: "Game created successfully"}, nil
}

func (s *GameServer) ListGames(ctx context.Context, _ *pb.ListGamesRequest) (*pb.ListGamesResponse, error) {
	games, err := s.games.ListGames(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbGames := make([]*pb.GameProto, 0, len(games))
	for _, g := range games {
		pbGames = append(pbGames, &pb.GameProto{
			Id:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			CreatedAt:   g.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.ListGamesResponse{Games: pbGames}, nil
}
