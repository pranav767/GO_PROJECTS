package controller

import (
	"e-commerce/model"
	"e-commerce/internal/service"
	"e-commerce/utils"
	"e-commerce/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	var users model.User
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
	var users model.User
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

// AddToCartHandler adds a product to the user's cart
func AddToCartHandler(c *gin.Context) {
    var req struct {
        ProductID uint `json:"product_id"`
        Quantity  int  `json:"quantity"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || req.Quantity <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    user, err := db.GetUserByUsername(username.(string))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }
    err = service.AddProduct(c.Request.Context(), req.ProductID, uint(user.ID), req.Quantity)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Product added to cart"})
}

// RemoveFromCartHandler removes a product or quantity from the user's cart
func RemoveFromCartHandler(c *gin.Context) {
    var req struct {
        ProductID uint `json:"product_id"`
        Quantity  int  `json:"quantity"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || req.Quantity <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    user, err := db.GetUserByUsername(username.(string))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }
    err = service.RemoveProduct(c.Request.Context(), req.ProductID, uint(user.ID), req.Quantity)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Product removed from cart"})
}

// GetCartHandler returns the user's cart
func GetCartHandler(c *gin.Context) {
    username, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    user, err := db.GetUserByUsername(username.(string))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }
    cart, err := service.GetCart(c.Request.Context(), uint(user.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, cart)
}