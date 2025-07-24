// Package models contains shared data structures
package models

import (
	"time"
)

// BlogPost represents a blog post entity
type BlogPost struct {
	ID        string    `json:"id" bson:"id_str,omitempty"`
	NumericID int       `json:"numeric_id" bson:"id"`
	Title     string    `json:"title" binding:"required" bson:"title"`
	Content   string    `json:"content" binding:"required" bson:"content"`
	Category  string    `json:"category" binding:"required" bson:"category"`
	Tags      []string  `json:"tags" binding:"required" bson:"tags"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"createdAt"`
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updatedAt,omitempty"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// NewSuccessResponse creates a standardized success response
func NewSuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}
