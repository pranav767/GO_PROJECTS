package controller

import (
	"leaderboard_system/internal/realtime"

	"github.com/gin-gonic/gin"
)

// WebSocketLeaderboardHandler delegates to realtime package handler
func WebSocketLeaderboardHandler(c *gin.Context) { realtime.Handler(c) }

// StartLeaderboardBroadcaster starts realtime broadcaster (should be called during app init)
func StartLeaderboardBroadcaster() { realtime.Start() }
