package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

const dsn = "admin:adminpass@tcp(db:3306)/e_commerce_orders?parseTime=true"

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
		log.Fatalf("failed to open orders db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping orders db: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Create order from user's cart
	r.POST("/orders", func(c *gin.Context) {
		// Authenticate
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth"})
			return
		}
		tokenStr := strings.TrimSpace(auth)
		if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
			tokenStr = strings.TrimSpace(tokenStr[7:])
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil })
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		username, _ := claims["sub"].(string)
		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Read cart JSON from carts DB (cross-db query)
		var itemsJSON []byte
		err = db.QueryRowContext(context.Background(), "SELECT items FROM e_commerce_carts.carts WHERE username = ?", username).Scan(&itemsJSON)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch cart"})
			return
		}
		var items []CartItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid cart format"})
			return
		}
		if len(items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
			return
		}

		// Calculate total by looking up product prices in e_commerce_products
		var totalCents int64 = 0
		for _, it := range items {
			var price float64
			err := db.QueryRowContext(context.Background(), "SELECT price FROM e_commerce_products.products WHERE id = ?", it.ProductID).Scan(&price)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product %d not found", it.ProductID)})
				return
			}
			totalCents += int64(price*100.0) * int64(it.Quantity)
		}

		// Insert order
		res, err := db.ExecContext(context.Background(), "INSERT INTO orders (username, amount_cents, currency, status, stripe_intent, created_at) VALUES (?, ?, ?, ?, ?, NOW())", username, totalCents, "usd", "pending", "pending")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
			return
		}
		oid, _ := res.LastInsertId()

		// Insert order_items
		for _, it := range items {
			_, err := db.ExecContext(context.Background(), "INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)", oid, it.ProductID, it.Quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save order items"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"order_id": oid, "amount_cents": totalCents, "currency": "usd"})
	})

	if err := r.Run(":8084"); err != nil {
		log.Fatalf("order service failed to run: %v", err)
	}
}
