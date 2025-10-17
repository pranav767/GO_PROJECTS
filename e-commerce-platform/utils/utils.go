package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func LoadSecret() []byte {
	godotenv.Load()
	secret := os.Getenv("HMAC_SECRET")
	return []byte(secret)
}

func GenerateHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func CompareHash(hashedpassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedpassword, password) == nil
}

// GenerateJWT creates a JWT for a given username
func GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := LoadSecret()
	return token.SignedString(secret)
}

// ValidateJWT parses and validates a JWT string and returns the username if valid
func ValidateJWT(tokenString string) (string, error) {
	secret := LoadSecret()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		username, ok := claims["username"].(string)
		if !ok {
			return "", fmt.Errorf("username not found in token")
		}
		return username, nil
	}
	return "", fmt.Errorf("invalid token claims")
}
