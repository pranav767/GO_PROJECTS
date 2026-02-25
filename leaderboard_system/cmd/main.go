package main

import (
	"log"
	"leaderboard_system/internal/controller"
	"leaderboard_system/internal/db"
	"leaderboard_system/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	   if err := db.InitDB(); err != nil {
		   log.Fatalf("Failed to initialize db %v", err)
	   }

	   if err := db.InitRedis(); err != nil {
		   log.Fatalf("Failed to initialize Redis: %v", err)
	   }

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	// Start realtime broadcaster for WebSocket updates
	controller.StartLeaderboardBroadcaster()

	log.Println("Server starting on :8080")
	log.Println("Endpoints available:")
	log.Println("- POST /register - User registration")
	log.Println("- POST /login - User login")
	log.Println("- GET /ws/leaderboard - Real-time leaderboard WebSocket")
	log.Println("- POST /api/games - Create game (admin only)")
	log.Println("- GET /api/games - List all games")
	log.Println("- POST /api/submit-score - Submit score for a game")
	log.Println("- GET /api/leaderboard - Get game leaderboard")
	log.Println("- GET /api/global-leaderboard - Get global leaderboard")
	log.Println("- GET /api/user-rank - Get user rank in game")
	log.Println("- GET /api/top-players - Get top players for period")
	log.Println("- GET /api/score-history - Get score history")

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
