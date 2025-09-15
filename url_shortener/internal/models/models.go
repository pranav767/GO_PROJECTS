// File for all models
package models

import (
	"time"
)

// URL Model
type URL struct {
	ID          int       `json:"id" bson:"id"`
	URL         string    `json:"url" bson:"url"`
	ShortCode   string    `json:"shortCode" bson:"shortCode"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	AccessCount int       `json:"accessCount" bson:"accessCount"`
}

// URL created
type urlCreated struct {
	URL string `json:"url"`
}

// updated URL
type urlUpdated struct {
	URL string `json:"url"`
}

// Error
type ErrorResponse struct {
	Error error `json:"error"`
}
