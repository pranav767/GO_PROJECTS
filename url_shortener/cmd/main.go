// Main application entry point
package main

import (
	"log"

	"url_shortner/internal/controllers"
	"url_shortner/internal/repository"
	"url_shortner/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize DB
	if err := repository.DbInit(); err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}

	storage := services.NewMongoStorage()
	server := controllers.NewServer(storage)

	r := gin.Default()
	r.POST("/shorten", server.HandleCreate)
	r.PUT("/update/:shortCode", server.HandleUpdate)
	r.DELETE("/delete/:shortCode", server.HandleDelete)
	r.GET("/stats/:shortCode", server.HandleStats)
	r.GET("/:shortCode", server.HandleRedirect)
	r.GET("/details/:shortCode", server.HandleGetDetails)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
