// Database access layer
package services

import (
	"context"
	"fmt"
	"time"
	"url_shortner/internal/models"
	"url_shortner/internal/repository"
	"url_shortner/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type URLRepository interface {
	Create(url *models.URL) error
	Update(shortCode string, url *models.URL) error
	Delete(shortCode string) error
	GetByShortCode(shortCode string) (*models.URL, error)
	IncrementAccessCount(shortCode string) error
}

// MongoStorage implements URLRepository
type MongoStorage struct {
	collection *mongo.Collection
}

// NewMongoStorage returns a MongoStorage instance
func NewMongoStorage() *MongoStorage {
	db := repository.GetDB()
	collection := db.Collection("Url_collection")
	return &MongoStorage{collection: collection}
}

// Implement all methods using MongoDB queries (InsertOne, UpdateOne, DeleteOne, FindOne, etc.)
func (m *MongoStorage) Create(url *models.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), repository.ConnectionTimeout)
	defer cancel()

	// Check if the ShortCode already exists
	var urldetails models.URL
	err := m.collection.FindOne(ctx, bson.M{"shortCode": url.ShortCode}).Decode(&urldetails)
	if err == nil {
		// ShortCode already exists
		return fmt.Errorf("Short Code for this url already exists")
	}

	// Generate new ShortCode if not provided
	if url.ShortCode == "" {
		shortCode, err := utils.GenerateShortCode(8)
		if err != nil {
			return err
		}
		url.ShortCode = shortCode
	}

	// Assign new ID
	nextID, err := repository.GetNextID()
	if err != nil {
		return err
	}
	url.ID = nextID
	url.CreatedAt = time.Now()
	url.UpdatedAt = time.Now()
	url.AccessCount = 0

	_, err = m.collection.InsertOne(ctx, url)
	return err
}

func (m *MongoStorage) Update(shortCode string, url *models.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), repository.ConnectionTimeout)
	defer cancel()

	filter := bson.M{"shortCode": shortCode}
	update := bson.M{"$set": bson.M{
		"url":       url.URL,
		"updatedAt": time.Now(),
	}}
	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ShortCode not found")
	}
	return nil
}

func (m *MongoStorage) Delete(shortCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), repository.ConnectionTimeout)
	defer cancel()

	filter := bson.M{"shortCode": shortCode}
	result, err := m.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("ShortCode not found")
	}
	return nil
}

func (m *MongoStorage) GetByShortCode(shortCode string) (*models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repository.ConnectionTimeout)
	defer cancel()

	var url models.URL
	err := m.collection.FindOne(ctx, bson.M{"shortCode": shortCode}).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("ShortCode not found")
		}
		return nil, err
	}
	return &url, nil
}

func (m *MongoStorage) IncrementAccessCount(shortCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), repository.ConnectionTimeout)
	defer cancel()

	filter := bson.M{"shortCode": shortCode}
	update := bson.M{"$inc": bson.M{"accessCount": 1}, "$set": bson.M{"updatedAt": time.Now()}}
	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ShortCode not found")
	}
	return nil
}
