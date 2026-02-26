package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcdelivery "leaderboard_system/internal/delivery/grpc"
	"leaderboard_system/internal/service"
)

// publicMethods are gRPC methods that do not require JWT auth.
var publicMethods = map[string]bool{
	"/leaderboard.v1.AuthService/Register": true,
	"/leaderboard.v1.AuthService/Login":    true,
	"/grpc.health.v1.Health/Check":         true,
	"/grpc.health.v1.Health/Watch":         true,
}

// AuthUnaryInterceptor creates a gRPC unary interceptor for JWT authentication.
func AuthUnaryInterceptor(authSvc *service.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(auth[0], "Bearer ")
		username, err := authSvc.ValidateJWT(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Look up user to get ID and role
		user, err := authSvc.GetUserByUsername(ctx, username)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "user not found")
		}

		ctx = context.WithValue(ctx, grpcdelivery.CtxKeyUsername, username)
		ctx = context.WithValue(ctx, grpcdelivery.CtxKeyUserID, user.ID)
		ctx = context.WithValue(ctx, grpcdelivery.CtxKeyRole, user.Role)

		return handler(ctx, req)
	}
}
