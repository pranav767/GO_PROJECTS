package handler

import (
	"blogging_platform/db"
	"blogging_platform/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateBlogPost handles the creation of a new blog post
func CreateBlogPost(c *gin.Context) {
	var newPost models.BlogPost

	// Parse request body
	if err := c.ShouldBindJSON(&newPost); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Validate required fields
	if !validateBlogPost(&newPost) {
		sendError(c, http.StatusBadRequest, "All fields are required")
		return
	}

	// Create context with timeout
	ctx, cancel := createContext()
	defer cancel()

	// Get database connection
	database := db.GetDB()
	if database == nil {
		sendError(c, http.StatusInternalServerError, "Database connection error")
		return
	}

	collection := database.Collection(db.BlogPostCollection)

	// Get next ID
	postID, err := db.GetNextID()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Error generating ID: "+err.Error())
		return
	}

	// Set post metadata
	newPost.NumericID = postID
	newPost.ID = fmt.Sprintf("%d", postID)
	newPost.CreatedAt = time.Now()

	// Insert the document
	result, err := collection.InsertOne(ctx, newPost)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create blog post: "+err.Error())
		return
	}

	// Return success response
	data := map[string]interface{}{
		"id":       postID,
		"mongo_id": result.InsertedID,
		"post":     newPost,
	}
	sendSuccess(c, http.StatusCreated, "Blog post created successfully", data)
}

// GetBlogPost retrieves all blog posts
func GetBlogPost(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := createContext()
	defer cancel()

	database := db.GetDB()
	if database == nil {
		sendError(c, http.StatusInternalServerError, "Database connection error")
		return
	}

	collection := database.Collection(db.BlogPostCollection)

	// Options to sort by ID in ascending order
	opts := options.Find().SetSort(bson.M{"id": 1})

	// Execute the query
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to retrieve blog posts: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	// Decode all documents into a slice
	var posts []models.BlogPost
	if err := cursor.All(ctx, &posts); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to decode blog posts: "+err.Error())
		return
	}

	// Return the posts as JSON
	data := map[string]interface{}{
		"count": len(posts),
		"posts": posts,
	}
	sendSuccess(c, http.StatusOK, "Blog posts retrieved successfully", data)
}

// DeleteBlogPost deletes a blog post by ID
func DeleteBlogPost(c *gin.Context) {
	// Get post ID from URL parameter
	postID := c.Param("id")
	if postID == "" {
		sendError(c, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Convert ID string to int for query
	id, err := strconv.Atoi(postID)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	// Create context with timeout
	ctx, cancel := createContext()
	defer cancel()

	// Get database connection
	database := db.GetDB()
	if database == nil {
		sendError(c, http.StatusInternalServerError, "Database connection error")
		return
	}

	collection := database.Collection(db.BlogPostCollection)

	// Delete the post with matching ID
	result, err := collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete post: "+err.Error())
		return
	}

	// Check if any document was deleted
	if result.DeletedCount == 0 {
		sendError(c, http.StatusNotFound, "No post found with that ID")
		return
	}

	// Return success response
	data := map[string]interface{}{
		"id": postID,
	}
	sendSuccess(c, http.StatusOK, "Blog post deleted successfully", data)
}

// UpdateBlogPost updates an existing blog post
func UpdateBlogPost(c *gin.Context) {
	// Get post ID from URL parameter
	postID := c.Param("id")
	if postID == "" {
		sendError(c, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Convert ID string to int for query
	id, err := strconv.Atoi(postID)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	// Parse the updated post data
	var updatedPost models.BlogPost
	if err := c.ShouldBindJSON(&updatedPost); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Validate required fields
	if !validateBlogPost(&updatedPost) {
		sendError(c, http.StatusBadRequest, "All fields are required")
		return
	}

	// Create context with timeout
	ctx, cancel := createContext()
	defer cancel()

	// Get database connection
	database := db.GetDB()
	if database == nil {
		sendError(c, http.StatusInternalServerError, "Database connection error")
		return
	}

	collection := database.Collection(db.BlogPostCollection)

	// First check if the post exists
	var existingPost models.BlogPost
	err = collection.FindOne(ctx, bson.M{"id": id}).Decode(&existingPost)
	if err != nil {
		sendError(c, http.StatusNotFound, "Post not found")
		return
	}

	// Create update document
	update := bson.M{
		"$set": bson.M{
			"title":     updatedPost.Title,
			"content":   updatedPost.Content,
			"category":  updatedPost.Category,
			"tags":      updatedPost.Tags,
			"updatedAt": time.Now(),
		},
	}

	// Update the document
	result, err := collection.UpdateOne(ctx, bson.M{"id": id}, update)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update post: "+err.Error())
		return
	}

	// Check if any document was updated
	if result.ModifiedCount == 0 {
		sendError(c, http.StatusNotFound, "No post was updated")
		return
	}

	// Set the ID in the response
	updatedPost.ID = postID
	updatedPost.NumericID = id

	// Return success response
	data := map[string]interface{}{
		"id":   id,
		"post": updatedPost,
	}
	sendSuccess(c, http.StatusOK, "Blog post updated successfully", data)
}

// Helper functions are defined in utils.go
