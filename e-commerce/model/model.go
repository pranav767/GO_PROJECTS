package model

import (
	"sync"
	//"gorm.io/gorm"
)

// userStore struct which will have mutex & users map to store username & passwd
type UserStore struct {
	Mu   sync.Mutex
	User map[string]string
}

type User struct {
	Username string
	Password string
}

type Product struct {
	ID          uint    `gorm:"primaryKey; autoincrement"  json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Inventory   int     `json:"inventory"`
}

type CartItem struct {
	ID        uint `gorm:"primaryKey;autoincrement"  json:"id"`
	CartID    uint `json:"cart_id"`
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type OrderItem struct {
	ID        uint `gorm:"primaryKey;autoincrement" json:"id"`
	OrderID   uint `json:"order_id"`
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type Cart struct {
	ID     uint       `gorm:"primaryKey;autoincrement" json:"id"`
	UserID uint       `json:"user_id"`
	Items  []CartItem `gorm:"foreignKey:CartID" json:"items"`
}

type PaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type Order struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"user_id"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	StripeIntent string  `json:"stripe_intent"`
}
