package service

import (
	"errors"
	"movie-reservation/internal/db"
	"movie-reservation/model"
	"movie-reservation/utils"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// Load secrets using godotenv
func loadSecret() []byte {
	godotenv.Load()
	secret := os.Getenv("HMAC_SECRET")
	return []byte(secret)
}

// Register new user
func RegisterUser(username string, password string) error {

	// Check if user already exists
	user, _ := db.GetUserByUserName(username)
	if user != nil {
		return errors.New("user already exists")
	}

	// Generate password hash
	hash, err := utils.GenerateHash([]byte(password))
	if err != nil {
		return errors.New("error while generating hash")
	}

	_, err = db.CreateUser(username, string(hash), "user")
	return err
}

func AuthenticateUser(username string, password string) (bool, error) {

	//Get user details
	user, err := db.GetUserByUserName(username)
	if err != nil {
		return false, errors.New("user not found")
	}

	if !utils.CompareHashwithPassword([]byte(user.Password), []byte(password)) {
		return false, errors.New("invalid passwd")
	}
	return true, nil
}

func GenerateJWT(username string) (string, error) {
	// New claim
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	// Sign and get the complete encoded token as a string using the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := loadSecret()
	return token.SignedString(secret)
}

// Validate JWT token
func ValidateJWT(tokenString string) (string, error) {
	secret := loadSecret()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("username not found in token")
		}
		return username, nil
	}
	return "", errors.New("invalid token")
}

// GetUserByUsername retrieves user by username
func GetUserByUsername(username string) (*model.User, error) {
	return db.GetUserByUserName(username)
}
