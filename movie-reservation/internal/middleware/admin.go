package middleware

import (
	"movie-reservation/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware checks if the authenticated user has admin role
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Get user details to check role
		user, err := db.GetUserByUserName(username.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		// Store user role in context for handlers
		c.Set("userRole", user.Role)
		c.Next()
	}
}

// RoleAuthMiddleware checks if user has any of the specified roles
func RoleAuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, err := db.GetUserByUserName(username.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		roleAllowed := false
		for _, role := range allowedRoles {
			if user.Role == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("userRole", user.Role)
		c.Next()
	}
}
