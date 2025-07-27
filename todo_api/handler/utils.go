// Package handler contains HTTP request handlers
package handler

import (
	"context"
	"time"
	"todo_api/models"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultTimeout for context operations
	DefaultTimeout = 5 * time.Second
)

// createContext creates a timeout context for database operations
func CreateContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), DefaultTimeout)
}

// sendError sends a standardized error response
func sendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.NewErrorResponse(message))
}

// sendSuccess sends a standardized success response
func sendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, models.NewSuccessResponse(message, data))
}