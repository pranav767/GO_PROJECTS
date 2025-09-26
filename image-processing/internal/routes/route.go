package routes

import (
	"image-processing/internal/controller"
	"image-processing/internal/middleware"
	"image-processing/internal/service"
	"image-processing/model"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, s3Client *model.S3Client) {
	// Public routes (no authentication required)
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)

	// Initialize services
	uploadService := service.NewImageUploadService(s3Client)
	processingService := service.NewImageProcessingService(s3Client)
	imageController := controller.NewImageController(uploadService, processingService)

	// Protected routes (JWT authentication required)
	api := r.Group("/")
	api.Use(middleware.JWTAuthMiddleware()) // Apply JWT middleware to this group
	{
		// Image management endpoints
		api.POST("/images", imageController.UploadImage)
		api.GET("/images", imageController.GetImages)
		api.GET("/images/:id", imageController.GetImage)
		api.DELETE("/images/:id", imageController.DeleteImage)

		// Image transformation endpoints
		api.POST("/images/:id/transform", imageController.TransformImage)
		api.GET("/images/:id/transformations", imageController.GetTransformations)
	}
}
