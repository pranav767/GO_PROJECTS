package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"leaderboard_system/internal/domain"
	"leaderboard_system/internal/service"
)

// mockUserRepo implements domain.UserRepository for testing.
type mockUserRepo struct {
	users  map[string]*domain.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) CreateUser(_ context.Context, username, passwordHash string) (int64, error) {
	m.nextID++
	m.users[username] = &domain.User{
		ID:           m.nextID,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         "user",
	}
	return m.nextID, nil
}

func (m *mockUserRepo) GetUserByUsername(_ context.Context, username string) (*domain.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) UpdateUserRole(_ context.Context, userID int64, role string) error {
	for _, u := range m.users {
		if u.ID == userID {
			u.Role = role
			return nil
		}
	}
	return domain.ErrUserNotFound
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	err := svc.Register(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify user was created
	u, err := repo.GetUserByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("user not found after registration: %v", err)
	}
	if u.Username != "alice" {
		t.Fatalf("expected username 'alice', got '%s'", u.Username)
	}
	if u.Role != "user" {
		t.Fatalf("expected role 'user', got '%s'", u.Role)
	}
}

func TestRegister_DuplicateUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	_ = svc.Register(context.Background(), "alice", "password123")
	err := svc.Register(context.Background(), "alice", "password456")
	if err != domain.ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	_ = svc.Register(context.Background(), "alice", "password123")

	ok, err := svc.Authenticate(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatal("expected authentication to succeed")
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	_ = svc.Register(context.Background(), "alice", "password123")

	ok, err := svc.Authenticate(context.Background(), "alice", "wrongpassword")
	if ok {
		t.Fatal("expected authentication to fail")
	}
	if err != domain.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	ok, err := svc.Authenticate(context.Background(), "nonexistent", "password123")
	if ok {
		t.Fatal("expected authentication to fail")
	}
	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGenerateJWT_ValidToken(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, []byte("test-secret"), "", testLogger())

	_ = svc.Register(context.Background(), "alice", "password123")

	token, err := svc.GenerateJWT(context.Background(), "alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Roundtrip: validate the token we just generated
	username, err := svc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate generated token: %v", err)
	}
	if username != "alice" {
		t.Fatalf("expected username 'alice', got '%s'", username)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), []byte("test-secret"), "", testLogger())

	_, err := svc.ValidateJWT("invalid.token.string")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
