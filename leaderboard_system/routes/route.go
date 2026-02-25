package routes

import (
	"leaderboard_system/internal/controller"
	"leaderboard_system/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Public routes
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)
	// WebSocket endpoint for real-time leaderboard
	r.GET("/ws/leaderboard", controller.WebSocketLeaderboardHandler)
	
	// Protected routes (JWT authentication required)
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	{
		// Admin-only routes
		api.POST("/games", middleware.AdminOnly(), controller.CreateGameHandler)
		
		// Regular authenticated user routes
		api.GET("/games", controller.ListGamesHandler)
		api.POST("/submit-score", controller.SubmitScoreHandler)
		api.GET("/score-history", controller.GetScoreHistoryHandler)
		api.GET("/leaderboard", controller.GetLeaderboardHandler)
		api.GET("/global-leaderboard", controller.GetGlobalLeaderboardHandler)
		api.GET("/user-rank", controller.GetUserRankHandler)
		api.GET("/top-players", controller.GetTopPlayersByPeriodHandler)
	}
}
