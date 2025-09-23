package main

import (
	"log"
	"image-processing/internal/db"
	"github.com/gin-gonic/gin"
	"image-processing/internal/routes"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initalize db %v", err)
	}
	r := gin.Default()

	routes.SetupRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}