package interceptors_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcdelivery "leaderboard_system/internal/delivery/grpc"
	"leaderboard_system/internal/delivery/grpc/interceptors"
	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

type testUserRepo struct {
	users map[string]*domain.User
}

func (r *testUserRepo) CreateUser(_ context.Context, username, hash string) (int64, error) {
	id := int64(len(r.users) + 1)
	r.users[username] = &domain.User{ID: id, Username: username, PasswordHash: hash, Role: "user"}
	return id, nil
}

func (r *testUserRepo) GetUserByUsername(_ context.Context, username string) (*domain.User, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *testUserRepo) GetUserByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *testUserRepo) UpdateUserRole(_ context.Context, userID int64, role string) error {
	for _, u := range r.users {
		if u.ID == userID {
			u.Role = role
			return nil
		}
	}
	return domain.ErrUserNotFound
}

func setupAuthTest(t *testing.T) (*service.AuthService, grpc.UnaryServerInterceptor) {
	t.Helper()
	repo := &testUserRepo{users: make(map[string]*domain.User)}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	authSvc := service.NewAuthService(repo, []byte("test-secret"), "", logger)

	// Register a test user
	_ = authSvc.Register(context.Background(), "testuser", "password123")

	interceptor := interceptors.AuthUnaryInterceptor(authSvc)
	return authSvc, interceptor
}

func TestAuthInterceptor_PublicMethodsPassThrough(t *testing.T) {
	_, interceptor := setupAuthTest(t)

	info := &grpc.UnaryServerInfo{FullMethod: "/leaderboard.v1.AuthService/Register"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("expected no error for public method, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	_, interceptor := setupAuthTest(t)

	info := &grpc.UnaryServerInfo{FullMethod: "/leaderboard.v1.LeaderboardService/SubmitScore"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	_, err := interceptor(context.Background(), nil, info, handler)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestAuthInterceptor_ValidToken_SetsContext(t *testing.T) {
	authSvc, interceptor := setupAuthTest(t)

	token, err := authSvc.GenerateJWT(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/leaderboard.v1.LeaderboardService/SubmitScore"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify context values were set
		username := ctx.Value(grpcdelivery.CtxKeyUsername)
		if username != "testuser" {
			t.Fatalf("expected username 'testuser', got %v", username)
		}
		userID := ctx.Value(grpcdelivery.CtxKeyUserID)
		if userID == nil {
			t.Fatal("expected userID in context")
		}
		role := ctx.Value(grpcdelivery.CtxKeyRole)
		if role != "user" {
			t.Fatalf("expected role 'user', got %v", role)
		}
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}
