package controller

import (
	//"image-processing/internal/db"
	"image-processing/internal/service"
	"image-processing/model"
	"net/http"

	"github.com/gin-gonic/gin"
	//"golang.org/x/text/cases"
)


func RegisterHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	err = service.RegisterUser(user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message":"User registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	exist, err := service.AuthenticateUser(user.Username, user.Password)
	if !exist{
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusInternalServerError, gin.H{"error":"User does not exist"})
			case "invalid passwd":
				c.JSON(http.StatusForbidden, gin.H{"error":"Invalid password"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error":"Authentication failed"})
			} 
		}	else {
				c.JSON(http.StatusUnauthorized, gin.H{"error":"Auth failed"})
		}
		return
		}
	token, err := service.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate token"})
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}