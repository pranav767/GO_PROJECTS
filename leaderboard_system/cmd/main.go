package main

import (
	"log"
	"leaderboard_system/internal/db"
	"leaderboard_system/routes"

	"github.com/gin-gonic/gin"
)

func main() {

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

	log.Println("Server starting on :8080")
	log.Println("Endpoints available:")
	log.Println("- POST /register - User registration")
	log.Println("- POST /login - User login")
	log.Println("- GET /api/exercises - List exercises (with optional filters)")
	log.Println("- POST /api/workout-plans - Create workout plan")
	log.Println("- GET /api/workout-plans - List user workout plans")
	log.Println("- POST /api/workout-sessions - Schedule a workout")
	log.Println("- GET /api/workout-sessions - List workout sessions")
	log.Println("- GET /api/reports/workout - Generate workout report")

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
