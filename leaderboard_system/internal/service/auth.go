package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"leaderboard_system/internal/domain"
)

// AuthService handles user authentication and JWT token management.
type AuthService struct {
	users         domain.UserRepository
	secret        []byte
	adminUsername string
	logger        *slog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(users domain.UserRepository, secret []byte, adminUsername string, logger *slog.Logger) *AuthService {
	return &AuthService{
		users:         users,
		secret:        secret,
		adminUsername: adminUsername,
		logger:        logger,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, username, password string) error {
	_, err := s.users.GetUserByUsername(ctx, username)
	if err == nil {
		return domain.ErrUserExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID, err := s.users.CreateUser(ctx, username, string(hash))
	if err != nil {
		return err
	}

	if s.adminUsername != "" && username == s.adminUsername {
		if err := s.users.UpdateUserRole(ctx, userID, "admin"); err != nil {
			s.logger.Warn("created user but failed to promote to admin",
				slog.String("username", username),
				slog.Any("error", err),
			)
		}
	}

	return nil
}

// Authenticate verifies credentials.
func (s *AuthService) Authenticate(ctx context.Context, username, password string) (bool, error) {
	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		return false, domain.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, domain.ErrInvalidPassword
	}

	return true, nil
}

// GenerateJWT creates a signed JWT for the given username.
func (s *AuthService) GenerateJWT(ctx context.Context, username string) (string, error) {
	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	claims := jwt.MapClaims{
		"username": username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateJWT parses and validates a JWT, returning the username from claims.
func (s *AuthService) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	return username, nil
}

// GetUserByUsername retrieves a user by username.
func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.users.GetUserByUsername(ctx, username)
}
