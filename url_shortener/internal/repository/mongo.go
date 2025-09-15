package repository

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DatabaseName      = "Url_db"
	CollectionName    = "Url_collection"
	ConnectionTimeout = 5 * time.Second
	OperationTimeout  = 5 * time.Second
)

// Global Variables
var (
	client *mongo.Client
	mutex  sync.Mutex
	once   sync.Once
)

// DB init
func DbInit() error {
	var initError error
	once.Do(func() {
		initError = initializeClient()
	})
	return initError
}

func initializeClient() error {
	// set a uri to connect to mongodb instance
	uri := "mongodb://admin:adminpass@localhost:27017"

	// This function returns an empty, non-cancelable context, which is commonly used as the root context for operations that don't have a specific parent context.
	//ctx,cancel := context.WithTimeout(Context.Background() , ConnectionTimeout)
	// Cancel is a function returned when used withTimeout or other methods
	//defer cancel()

	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Error while connecting to DB")
		return err
	}
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
	collection := client.Database(DatabaseName).Collection(CollectionName)

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

// Get DB returns DB instance

func GetDB() *mongo.Database {
	if client == nil {
		log.Printf("Warning : Database client not initialized")
		return nil
	}
	return client.Database(DatabaseName)
}

// Close disconnects the mongodb instance
func Close() error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), ConnectionTimeout)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect from DB")
		return err
	}
	log.Printf("Successfully disconnected from mongodb")
	return nil
}
