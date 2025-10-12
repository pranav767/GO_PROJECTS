package controller

import (
	"net/http"
	"strconv"
	"leaderboard_system/internal/service"
	"github.com/gin-gonic/gin"
)

// GetScoreHistoryHandler returns a user's score history for a game
func GetScoreHistoryHandler(c *gin.Context) {
	userIDStr := c.Query("user_id")
	game := c.Query("game")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}
	scores, err := service.GetScoreHistoryService(userID, game)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scores)
}
