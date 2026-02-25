package controller

import (
	"net/http"
	"strconv"
	"leaderboard_system/internal/service"
	"github.com/gin-gonic/gin"
)

// GetScoreHistoryHandler returns a user's score history for a game
func GetScoreHistoryHandler(c *gin.Context) {
	game := c.Query("game")
	role := c.GetString("role")
	authenticatedUserIDStr := c.GetString("userID")
	
	// Determine which user's history to retrieve
	var targetUserID int64
	var err error
	
	// Admins can query any user via user_id param, regular users can only see their own
	if role == "admin" {
		userIDStr := c.Query("user_id")
		if userIDStr != "" {
			targetUserID, err = strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
				return
			}
		} else {
			// Admin querying their own history
			targetUserID, err = strconv.ParseInt(authenticatedUserIDStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authenticated user_id"})
				return
			}
		}
	} else {
		// Regular users can only see their own history
		targetUserID, err = strconv.ParseInt(authenticatedUserIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
	}
	
	scores, err := service.GetScoreHistoryService(c.Request.Context(), targetUserID, game)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scores)
}
