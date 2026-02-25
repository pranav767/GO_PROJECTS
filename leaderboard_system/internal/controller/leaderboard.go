package controller

import (
	"net/http"
	"strconv"
	"time"

	"leaderboard_system/internal/service"

	"github.com/gin-gonic/gin"
)

// SubmitScoreHandler handles score submissions
func SubmitScoreHandler(c *gin.Context) {
	var req struct {
		Game  string  `json:"game"`
		Score float64 `json:"score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authenticated user"})
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	if err := service.SubmitScoreService(c.Request.Context(), userID, req.Game, req.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Score submitted"})
}

// GetLeaderboardHandler returns the top N leaderboard for a game
func GetLeaderboardHandler(c *gin.Context) {
	game := c.Query("game")
	topNStr := c.DefaultQuery("topN", "10")
	topN, _ := strconv.ParseInt(topNStr, 10, 64)
	if topN <= 0 {
		topN = 10
	}
	if topN > 100 {
		topN = 100
	}
	entries, err := service.GetLeaderboardService(c.Request.Context(), game, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GetUserRankHandler returns a user's rank for a game
func GetUserRankHandler(c *gin.Context) {
	game := c.Query("game")
	userIDStr := c.GetString("userID")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	entry, err := service.GetUserRankService(c.Request.Context(), game, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
}

// GetTopPlayersByPeriodHandler returns top N players for a game in a period
func GetTopPlayersByPeriodHandler(c *gin.Context) {
	game := c.Query("game")
	if game == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is required"})
		return
	}
	periodStr := c.Query("period") // optional: defaults to today
	if periodStr == "" {
		periodStr = time.Now().Format("2006-01-02")
	}
	topNStr := c.DefaultQuery("topN", "10")
	topN, _ := strconv.ParseInt(topNStr, 10, 64)
	if topN <= 0 {
		topN = 10
	}
	if topN > 100 {
		topN = 100
	}
	period, err := time.Parse("2006-01-02", periodStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period format, expected YYYY-MM-DD"})
		return
	}
	entries, err := service.GetTopPlayersByPeriodService(c.Request.Context(), game, period, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": game, "period": periodStr, "topN": topN, "entries": entries})
}
