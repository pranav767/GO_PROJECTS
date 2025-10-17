package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const dsn = "admin:adminpass@tcp(db:3306)/e_commerce_orders?parseTime=true"

func main() {
	// For MVP we'll simulate payment intents
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

	// Create payment intent (simulate)
	r.POST("/create_intent", func(c *gin.Context) {
		var req struct{ OrderID int64 }
		if err := c.BindJSON(&req); err != nil || req.OrderID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		// Lookup order amount
		var amount int64
		err := db.QueryRowContext(context.Background(), "SELECT amount_cents FROM orders WHERE id = ?", req.OrderID).Scan(&amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
			return
		}
		// Simulate a PaymentIntent ID and client_secret
		piID := "pi_" + strings.ReplaceAll(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(strings.Repeat("X", 8)), " ", "")), "", "")
		clientSecret := "cs_test_" + piID

		// Update order stripe_intent
		_, _ = db.ExecContext(context.Background(), "UPDATE orders SET stripe_intent = ? WHERE id = ?", piID, req.OrderID)

		c.JSON(http.StatusOK, gin.H{"payment_intent_id": piID, "client_secret": clientSecret, "amount": amount})
	})

	// Webhook endpoint to mark order paid (simulated)
	r.POST("/webhook", func(c *gin.Context) {
		var ev struct {
			Type string
			Data map[string]interface{}
		}
		if err := c.BindJSON(&ev); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
			return
		}
		if ev.Type == "payment_intent.succeeded" {
			if idv, ok := ev.Data["id"].(string); ok {
				// Mark order paid by stripe_intent
				_, err := db.ExecContext(context.Background(), "UPDATE orders SET status = 'paid' WHERE stripe_intent = ?", idv)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
					return
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "received"})
	})

	if err := r.Run(":8085"); err != nil {
		log.Fatalf("payment service failed to run: %v", err)
	}
}
