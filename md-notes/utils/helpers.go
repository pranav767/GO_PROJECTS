package utils

import "github.com/google/uuid"

// GenerateID generates a new unique ID for notes.
func GenerateID() string {
    return uuid.New().String()
}