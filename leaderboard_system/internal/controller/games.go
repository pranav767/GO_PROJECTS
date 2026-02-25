package controller

import (
	"log"
	"net/http"
	"strings"

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
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game name is required"})
		return
	}
	log.Printf("[Game] Create game request: name=%s", req.Name)
	id, err := service.CreateGameService(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		log.Printf("[Game] Create game failed for name=%s: %v", req.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[Game] Game created: id=%d name=%s", id, req.Name)
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "Game created"})
}

// ListGamesHandler lists all games
func ListGamesHandler(c *gin.Context) {
	log.Printf("[Game] List games request")
	games, err := service.ListGamesService(c.Request.Context())
	if err != nil {
		log.Printf("[Game] List games failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, games)
}
