package routes

import (
	"e-commerce/internal/controller"
	"e-commerce/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all HTTP routes for the application
func SetupRoutes(r *gin.Engine) {
	r.POST("/signup", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)
	r.POST("/webhook", controller.StripeWebhookHandler)

	// Protected route group
	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	auth.GET("/profile", controller.ProfileHandler)
    auth.POST("/cart/add", controller.AddToCartHandler)
    auth.POST("/cart/remove", controller.RemoveFromCartHandler)
    auth.GET("/cart", controller.GetCartHandler)
	auth.POST("/checkout", controller.CreatePaymentIntentHandler)
}
