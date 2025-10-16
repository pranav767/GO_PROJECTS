package db

import (
	"context"
	"e-commerce/model"
	"encoding/json"
)

func GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	var p model.Product
	err := db.QueryRowContext(ctx, "SELECT id, name, description, price, inventory FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Inventory)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Get user's cart details by userID
func GetUserCartDetailsByUserID(ctx context.Context, userID uint) (*model.Cart, error) {
	var user model.Cart
	var itemsData []byte
	err := db.QueryRowContext(ctx, "SELECT id, userid, items FROM cart WHERE userid = ?", userID).Scan(&user.ID, &user.UserID, &itemsData)
	if err != nil {
		return nil, err
	}
	// MySQL returns the JSON column as []byte, not as a Go struct/slice.
	// We must unmarshal the JSON bytes into the Go slice type ([]model.CartItem).
	if err := json.Unmarshal(itemsData, &user.Items); err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user's cart details by userID
func UpdateUserCartDetailsByUserID(ctx context.Context, cart *model.Cart, userID uint) error {
	// Serialize cart.Items to JSON
	itemsJSON, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		"UPDATE cart SET items = ? WHERE userid = ?",
		string(itemsJSON), userID,
	)
	return err
}

// CreateCartForUser creates a new cart for the user
func CreateCartForUser(ctx context.Context, cart *model.Cart) error {
	itemsJSON, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO cart (userid, items) VALUES (?, ?)", cart.UserID, string(itemsJSON))
	return err
}
