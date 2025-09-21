package main

import (
	"e-commerce/internal/db"
	"e-commerce/internal/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// Register all routes in a separate function
	routes.SetupRoutes(r)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
