package controller

import (
	"net/http"

	"movie-reservation/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := service.RegisterUser(req.Username, req.Password); err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	ok, err := service.AuthenticateUser(req.Username, req.Password)
	if !ok {
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
			case "invalid passwd":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		}
		return
	}
	token, err := service.GenerateJWT(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
