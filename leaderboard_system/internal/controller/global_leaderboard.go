package controller

import (
	"net/http"
	"strconv"
	"leaderboard_system/internal/service"
	"github.com/gin-gonic/gin"
)

// GetGlobalLeaderboardHandler returns the global leaderboard
func GetGlobalLeaderboardHandler(c *gin.Context) {
	topNStr := c.DefaultQuery("topN", "10")
	topN, _ := strconv.ParseInt(topNStr, 10, 64)
	if topN <= 0 {
		topN = 10
	}
	if topN > 100 {
		topN = 100
	}
	entries, err := service.GetGlobalLeaderboardService(c.Request.Context(), topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}
