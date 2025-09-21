package service

import (
	"context"
	"database/sql"
	"e-commerce/internal/db"
	"e-commerce/model"
	"errors"
)

// Add to Cart for a User
func AddProduct(ctx context.Context, ProductID uint, userID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	p, err := db.GetProductByID(ctx, int(ProductID))
	if err != nil {
		return err
	}
	if p.Inventory < quantity {
		return errors.New("not enough Product Available")
	}
	// Not reducing the inventory, will do at checkout
	cart, err := db.GetUserCartDetailsByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create a new cart for this user
			cart = &model.Cart{UserID: userID, Items: []model.CartItem{}}
			if err := db.CreateCartForUser(ctx, cart); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Check if product already in cart
	found := false
	for i, item := range cart.Items {
		if item.ProductID == ProductID {
			if cart.Items[i].Quantity+quantity > p.Inventory {
				return errors.New("not enough Product Available")
			}
			cart.Items[i].Quantity += quantity
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, model.CartItem{
			ProductID: ProductID,
			Quantity:  quantity,
		})
	}

	// Update cart in DB
	err = db.UpdateUserCartDetailsByUserID(ctx, cart, userID)
	if err != nil {
		return err
	}
	return nil
}

// GetCart returns the user's cart
func GetCart(ctx context.Context, userID uint) (*model.Cart, error) {
	return db.GetUserCartDetailsByUserID(ctx, userID)
}

// Remove from Cart for a user
func RemoveProduct(ctx context.Context, productID uint, userID uint, quantity int) error {
	// Get UserCart Details
	cart, err := db.GetUserCartDetailsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for i, item := range cart.Items {
		if item.ProductID == productID {
			// If Available is more than to be removed
			if item.Quantity > quantity {
				cart.Items[i].Quantity -= quantity
			} else {
				// remove entire item
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			}
			return db.UpdateUserCartDetailsByUserID(ctx, cart, userID)
		}
	}
	return errors.New("product not found in cart")
}
