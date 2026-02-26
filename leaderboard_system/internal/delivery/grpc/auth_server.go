package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "leaderboard_system/api/gen/leaderboard/v1"
	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

// AuthServer implements the AuthService gRPC server.
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	auth *service.AuthService
}

// NewAuthServer creates a new AuthServer.
func NewAuthServer(auth *service.AuthService) *AuthServer {
	return &AuthServer{auth: auth}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := s.auth.Register(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ok, err := s.auth.Authenticate(ctx, req.Username, req.Password)
	if !ok {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user does not exist")
		}
		if errors.Is(err, domain.ErrInvalidPassword) {
			return nil, status.Error(codes.Unauthenticated, "invalid password")
		}
		return nil, status.Error(codes.Internal, "authentication failed")
	}

	token, err := s.auth.GenerateJWT(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, _ *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	return &pb.GetProfileResponse{
		Id:       userID,
		Username: usernameFromContext(ctx),
		Role:     roleFromContext(ctx),
	}, nil
}
