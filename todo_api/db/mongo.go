// Package db provides MongoDB database connectivity and operations
package db

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Constants for configuration
const (
	UserCollection = "users"
	TodoCollection = "todo"
	DatabaseName   = "todo_db"
	//BlogPostCollection = "todo_posts"
	ConnectTimeout   = 5 * time.Second
	OperationTimeout = 5 * time.Second
)

// Global variables
var (
	client *mongo.Client
	mutex  sync.Mutex
	once   sync.Once
)

// DbInit initializes the MongoDB connection
// This should be called once at application startup
func DbInit() error {
	var initErr error
	// Use sync.Once to ensure this only runs once
	once.Do(func() {
		initErr = initializeClient()
	})
	return initErr
}

// initializeClient creates and configures the MongoDB client
func initializeClient() error {
	uri := "mongodb://admin:adminpass@localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer cancel()

	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return err
	}

	log.Printf("Successfully connected to MongoDB")
	return nil
}

// GetNextID returns a unique incremental ID by finding the max existing ID and adding 1
func GetNextID() (int, error) {
	// Use mutex to prevent race conditions if multiple requests come in simultaneously
	mutex.Lock()
	defer mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	if client == nil {
		return 0, errors.New("database not initialized")
	}

	// Get the blog_posts collection
	collection := client.Database(DatabaseName).Collection(TodoCollection)

	// Find the document with the highest ID
	// Sort by ID in descending order and get the first document
	opts := options.FindOne().SetSort(bson.M{"id": -1})
	var result bson.M

	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		// If no documents exist yet, start with ID 1
		return 1, nil
	}

	// Get the highest ID
	highestID := 0
	if idVal, exists := result["id"]; exists {
		// Convert to int if possible
		if idInt, ok := idVal.(int32); ok {
			highestID = int(idInt)
		} else if idInt, ok := idVal.(int64); ok {
			highestID = int(idInt)
		} else if idInt, ok := idVal.(int); ok {
			highestID = idInt
		}
	}

	// Return highest ID + 1
	return highestID + 1, nil
}

// GetDB returns the database instance
func GetDB() *mongo.Database {
	if client == nil {
		log.Printf("Warning: Database client is nil. Make sure DbInit() was called.")
		return nil
	}
	return client.Database(DatabaseName)
}

// Close gracefully disconnects from MongoDB
func Close() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect from MongoDB: %v", err)
		return err
	}

	log.Printf("Successfully disconnected from MongoDB")
	return nil
}
