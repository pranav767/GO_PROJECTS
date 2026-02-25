package controller

import (
	"log"
	"net/http"
	"strings"

	"leaderboard_system/internal/db"
	"leaderboard_system/internal/service"
	"leaderboard_system/model"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	if strings.TrimSpace(user.Username) == "" || len(user.Username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be at least 3 characters"})
		return
	}
	if len(user.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}
	log.Printf("[Auth] Register attempt for username: %s", user.Username)
	if err := service.RegisterUser(user.Username, user.Password); err != nil {
		log.Printf("[Auth] Register failed for %s: %v", user.Username, err)
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[Auth] User registered successfully: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	log.Printf("[Auth] Login attempt for username: %s", user.Username)
	exist, err := service.AuthenticateUser(user.Username, user.Password)
	if !exist {
		log.Printf("[Auth] Login failed for %s: %v", user.Username, err)
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
			case "invalid passwd":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth failed"})
		}
		return
	}
	token, err := service.GenerateJWT(user.Username)
	if err != nil {
		log.Printf("[Auth] JWT generation failed for %s: %v", user.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	log.Printf("[Auth] Login successful for username: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetProfileHandler returns the authenticated user's info from the JWT context.
func GetProfileHandler(c *gin.Context) {
	log.Printf("[Auth] Profile request for username: %s", c.GetString("username"))
	c.JSON(http.StatusOK, gin.H{
		"id":       c.GetString("userID"),
		"username": c.GetString("username"),
		"role":     c.GetString("role"),
	})
}

// HealthHandler checks MySQL and Redis connectivity.
func HealthHandler(c *gin.Context) {
	status := gin.H{}
	httpStatus := http.StatusOK

	sqlDB := db.GetDB()
	if sqlDB == nil || sqlDB.PingContext(c.Request.Context()) != nil {
		status["mysql"] = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["mysql"] = "ok"
	}

	if db.RedisClient == nil {
		status["redis"] = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else if err := db.RedisClient.Ping(c.Request.Context()).Err(); err != nil {
		log.Printf("[Health] Redis ping failed: %v", err)
		status["redis"] = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["redis"] = "ok"
	}

	c.JSON(httpStatus, status)
}
