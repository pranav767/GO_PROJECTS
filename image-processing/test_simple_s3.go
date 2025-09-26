package main

import (
	"context"
	"fmt"
	"log"

	"image-processing/internal/db"
)

func main() {
	fmt.Println("Initializing S3 connection...")

	// Initialize S3 connection
	s3Config, err := db.InitS3Connection()
	if err != nil {
		log.Fatalf("Failed to initialize S3: %v", err)
	}

	fmt.Printf("S3 Client created successfully!\n")
	fmt.Printf("Bucket: %s\n", s3Config.Bucket)
	fmt.Printf("Region: %s\n", s3Config.Region)

	// Test the connection using helper function
	ctx := context.Background()
	err = db.TestS3Connection(ctx, s3Config)
	if err != nil {
		log.Fatalf("S3 connection test failed: %v", err)
	}

	fmt.Println("S3 connection test passed!")
}
