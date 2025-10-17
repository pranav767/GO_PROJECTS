package controller

import (
	"e-commerce/internal/db"
	"e-commerce/internal/service"
	"e-commerce/model"
	"e-commerce/utils"
	"encoding/json"
	"net/http"
	"os"

	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
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

func CreatePaymentIntentHandler(c *gin.Context) {
	//https://docs.stripe.com/api/payment_intents
	// Get username from JWT context
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

	// Fetch user's cart
	cart, err := service.GetCart(c.Request.Context(), uint(user.ID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not fetch cart: " + err.Error()})
		return
	}
	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	// Calculate total from cart items and DB product prices
	var total int64 = 0
	var currency string = "usd" // Default, or fetch from product if needed
	for _, item := range cart.Items {
		product, err := db.GetProductByID(c.Request.Context(), int(item.ProductID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found: " + err.Error()})
			return
		}
		total += int64(product.Price * float64(item.Quantity) * 100) // price in cents
		// Optionally: currency = product.Currency
	}
	if total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart total is zero"})
		return
	}

	dbConn := db.GetDB()
	tx, err := dbConn.Begin()
	if err != nil {
		log.Printf("[Checkout] Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Insert order with a temporary PaymentIntent ID (will update after PaymentIntent creation)
	res, err := tx.Exec(
		"INSERT INTO orders (user_id, amount, currency, status, stripe_intent, created_at) VALUES (?, ?, ?, ?, ?, NOW())",
		user.ID, total, currency, "pending", "pending",
	)
	if err != nil {
		tx.Rollback()
		log.Printf("[Checkout] Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}
	orderID, _ := res.LastInsertId()

	// Insert each cart item into order_items
	for _, item := range cart.Items {
		_, err := tx.Exec(
			"INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
			orderID, item.ProductID, item.Quantity,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[Checkout] Failed to insert order item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order items"})
			return
		}
	}

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(total),
		Currency:           stripe.String(currency),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		tx.Exec("DELETE FROM orders WHERE id = ?", orderID) // Clean up order if payment fails
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update order with real PaymentIntent ID
	_, err = tx.Exec("UPDATE orders SET stripe_intent = ? WHERE id = ?", pi.ID, orderID)
	if err != nil {
		tx.Exec("DELETE FROM orders WHERE id = ?", orderID)
		tx.Rollback()
		log.Printf("[Checkout] Failed to update order with PaymentIntent ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order with PaymentIntent ID"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Checkout] Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit order transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client_secret": pi.ClientSecret, "payment_intent_id": pi.ID, "amount": total, "currency": currency})
}

func StripeWebhookHandler(c *gin.Context) {
	// https://docs.stripe.com/webhooks
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[Webhook] Failed to read request body: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Request too large"})
		return
	}
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Printf("[Webhook] STRIPE_WEBHOOK_SECRET is not set! Check your environment variables or .env file.")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook secret not configured"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if sigHeader == "" {
		log.Printf("[Webhook] Missing Stripe-Signature header.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		payload, sigHeader, endpointSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		log.Printf("⚠️  Webhook signature verification failed. Error: %v | Payload: %s | Header: %s | Secret: %s", err, string(payload), sigHeader, endpointSecret[:6]+"...")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Signature"})
		return
	}

	// Handle the event
	log.Printf("[Webhook] Received event type: %s", event.Type)
	switch event.Type {
	case "payment_intent.succeeded":
		// 1. Parse the PaymentIntent Object from the event
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("[Webhook] Failed to parse PaymentIntent: %v | Raw: %s", err, string(event.Data.Raw))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to Parse PaymentIntent"})
			return
		}
		log.Printf("[Webhook] Successful payment: %s", pi.ID)

		// 2. Find the order in DB by stripe PaymentIntent ID
		order, err := db.GetOrderByStripeIntentID(pi.ID)
		if err != nil {
			log.Printf("[Webhook] Order not found for PaymentIntentID: %s | Error: %v", pi.ID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		// 3. Verify amount matches before marking as paid
		orderAmountCents := int64(order.Amount)
		if order.Amount < 1000 { // If stored as dollars, convert to cents
			orderAmountCents = int64(order.Amount * 100)
		}
		if pi.Amount != orderAmountCents {
			log.Printf("[Webhook] PaymentIntent amount %d does not match order amount %d for order %d", pi.Amount, orderAmountCents, order.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment amount does not match order"})
			return
		}

		// 4. Mark order as paid
		err = db.MarkOrderAsPaid(order.ID)
		if err != nil {
			log.Printf("[Webhook] Failed to mark order as Paid: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
			return
		}

		// 5. Update the table, remove the product quantity from inventory
		err = db.UpdateInventoryAndCart(order.ID)
		if err != nil {
			log.Printf("[Webhook] Failed to update inventory/cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory/cart"})
			return
		}
		log.Printf("[Webhook] Order %d marked as paid and inventory updated.", order.ID)

	case "payment_intent.failed":
		log.Printf("[Webhook] PaymentIntent Failed.")
	default:
		log.Printf("[Webhook] Unhandled event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
