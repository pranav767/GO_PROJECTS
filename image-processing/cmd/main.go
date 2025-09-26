package main

import (
	"image-processing/internal/db"
	"image-processing/internal/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize db %v", err)
	}

	// Initialize S3 client
	s3Client, err := db.InitS3Connection()
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup routes with S3 client
	routes.SetupRoutes(r, s3Client)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
