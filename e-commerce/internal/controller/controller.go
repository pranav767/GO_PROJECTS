package controller

import (
	"e-commerce/internal/service"
	"e-commerce/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	var users service.User
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request format"})
		return
	}
	err := service.RegisterUser(users.Username, users.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var users service.User
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request format"})
		return
	}
	ok, err := service.AuthenticateUser(users.Username, users.Password)
	if !ok {
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
			case "invalid password":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		}
		return
	}
	token, err := utils.GenerateJWT(users.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func ProfileHandler(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user in context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": username})
}