package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

const dsn = "admin:adminpass@tcp(db:3306)/e_commerce_carts?parseTime=true"

type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func main() {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("dev-secret")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Get cart for a user (authenticated via JWT)
	r.GET("/cart", func(c *gin.Context) {
		username, ok := getUsernameFromRequest(c, jwtSecret)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		var itemsJSON []byte
		err := db.QueryRowContext(context.Background(), "SELECT items FROM carts WHERE username = ?", username).Scan(&itemsJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"items": []CartItem{}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var items []CartItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	r.POST("/cart/add", func(c *gin.Context) {
		username, ok := getUsernameFromRequest(c, jwtSecret)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		var req CartItem
		if err := c.BindJSON(&req); err != nil || req.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		// Load existing
		var items []CartItem
		var itemsJSON []byte
		err := db.QueryRowContext(context.Background(), "SELECT items FROM carts WHERE username = ?", username).Scan(&itemsJSON)
		if err == nil {
			_ = json.Unmarshal(itemsJSON, &items)
		}
		// merge
		found := false
		for i := range items {
			if items[i].ProductID == req.ProductID {
				items[i].Quantity += req.Quantity
				found = true
				break
			}
		}
		if !found {
			items = append(items, req)
		}
		newJSON, _ := json.Marshal(items)
		// upsert
		_, err = db.ExecContext(context.Background(), "INSERT INTO carts (username, items) VALUES (?, ?) ON DUPLICATE KEY UPDATE items = ?", username, string(newJSON), string(newJSON))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "added"})
	})

	r.POST("/cart/remove", func(c *gin.Context) {
		username, ok := getUsernameFromRequest(c, jwtSecret)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		var req CartItem
		if err := c.BindJSON(&req); err != nil || req.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		var items []CartItem
		var itemsJSON []byte
		err := db.QueryRowContext(context.Background(), "SELECT items FROM carts WHERE username = ?", username).Scan(&itemsJSON)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no cart"})
			return
		}
		_ = json.Unmarshal(itemsJSON, &items)
		for i := range items {
			if items[i].ProductID == req.ProductID {
				if items[i].Quantity > req.Quantity {
					items[i].Quantity -= req.Quantity
				} else {
					items = append(items[:i], items[i+1:]...)
				}
				break
			}
		}
		newJSON, _ := json.Marshal(items)
		_, err = db.ExecContext(context.Background(), "UPDATE carts SET items = ? WHERE username = ?", string(newJSON), username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})

	if err := r.Run(":8083"); err != nil {
		log.Fatalf("cart service failed to run: %v", err)
	}
}

// getUsernameFromRequest extracts and verifies a JWT from Authorization header.
// Accepts header formats: "Bearer <token>" or raw token.
func getUsernameFromRequest(c *gin.Context, secret []byte) (string, bool) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", false
	}
	tokenStr := strings.TrimSpace(auth)
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = strings.TrimSpace(tokenStr[7:])
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", false
	}
	return sub, true
}
