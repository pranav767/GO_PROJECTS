package db

import (
	"context"
	"e-commerce/model"
	"log"
)

// GetOrderByStripeIntentID finds an order by its Stripe PaymentIntent ID
func GetOrderByStripeIntentID(paymentIntentID string) (*model.Order, error) {
	var order model.Order
	err := db.QueryRow("SELECT id, user_id, amount, currency, status, stripe_intent FROM orders WHERE stripe_intent = ?", paymentIntentID).
		Scan(&order.ID, &order.UserID, &order.Amount, &order.Currency, &order.Status, &order.StripeIntent)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// MarkOrderAsPaid updates the order status to 'paid'
func MarkOrderAsPaid(orderID int64) error {
	_, err := db.Exec("UPDATE orders SET status = 'paid' WHERE id = ?", orderID)
	return err
}

// UpdateInventoryAndCart updates inventory and clears the user's cart for the order
func UpdateInventoryAndCart(orderID int64) error {
	// 1. Get the order and user_id
	var userID int64
	err := db.QueryRow("SELECT user_id FROM orders WHERE id = ?", orderID).Scan(&userID)
	if err != nil {
		return err
	}

	// 2. Get order items for this order
	rows, err := db.Query("SELECT product_id, quantity FROM order_items WHERE order_id = ?", orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int
		var quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			return err
		}
		_, err := db.Exec("UPDATE products SET inventory = inventory - ? WHERE id = ? AND inventory >= ?", quantity, productID, quantity)
		if err != nil {
			return err
		}
	}

	// 3. Clear the user's cart
	cart, err := GetUserCartDetailsByUserID(context.Background(), uint(userID))
	if err == nil {
		cart.Items = []model.CartItem{}
		_ = UpdateUserCartDetailsByUserID(context.Background(), cart, uint(userID))
	}

	log.Printf("Inventory decremented and cart cleared for order %d (user %d)", orderID, userID)
	return nil
}
