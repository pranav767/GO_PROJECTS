package routes

import (
	"image-processing/internal/controller"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine){
	r.POST("/signup", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)
}