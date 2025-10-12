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
		api.POST("/games", controller.CreateGameHandler)
		api.GET("/games", controller.ListGamesHandler)
		api.POST("/submit-score", controller.SubmitScoreHandler)
		api.GET("/score-history", controller.GetScoreHistoryHandler)
		api.GET("/leaderboard", controller.GetLeaderboardHandler)
		api.GET("/global-leaderboard", controller.GetGlobalLeaderboardHandler)
		api.GET("/user-rank", controller.GetUserRankHandler)
		api.GET("/top-players", controller.GetTopPlayersByPeriodHandler)
	}
	//	api.GET("/reports/workout", controller.GenerateWorkoutReportHandler)
	//	api.GET("/reports/personal-records", controller.GetPersonalRecordsHandler)
	//	api.GET("/reports/streaks", controller.GetWorkoutStreaksHandler)
	//}
}
