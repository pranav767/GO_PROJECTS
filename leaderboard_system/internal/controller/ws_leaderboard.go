package controller

import (
	"log"

	"leaderboard_system/internal/realtime"

	"github.com/gin-gonic/gin"
)

// WebSocketLeaderboardHandler delegates to realtime package handler
func WebSocketLeaderboardHandler(c *gin.Context) {
	log.Printf("[WS] WebSocket connection from %s", c.ClientIP())
	realtime.Handler(c)
}

// StartLeaderboardBroadcaster starts realtime broadcaster (should be called during app init)
func StartLeaderboardBroadcaster() { realtime.Start() }
