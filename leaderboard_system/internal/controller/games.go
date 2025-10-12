package controller

import (
	"net/http"
	"leaderboard_system/internal/service"
	"github.com/gin-gonic/gin"
)

// CreateGameHandler handles game creation
func CreateGameHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	id, err := service.CreateGameService(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Game created"})
}

// ListGamesHandler lists all games
func ListGamesHandler(c *gin.Context) {
	games, err := service.ListGamesService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, games)
}
