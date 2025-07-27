package handler

import (
	"net/http"
	"strconv"
	"todo_api/db"
	"todo_api/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func CreateTodo(c *gin.Context) {
	// Parse the JSON request body directly into the Todo struct
	var todo models.Todo
	if err := c.ShouldBindBodyWith(&todo, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid JSON format: "+err.Error()))
		return
	}
	// Retrieve username from context
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Username not found in context"))
		return
	}
	// Set the username in the Todo struct
	todo.Username = username

	// Set ID to a new ObjectID
	id, err := db.GetNextID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to generate ID: "+err.Error()))
		return
	}
	todo.ID = strconv.Itoa(id)

	collections := db.GetDB().Collection(db.TodoCollection)
	// Insert the Todo into the database
	ctx, cancel := CreateContext()
	defer cancel()
	_, err = collections.InsertOne(ctx, todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create todo: "+err.Error()))
		return
	}
	// Return success response
	c.JSON(http.StatusOK, models.NewSuccessResponse("Todo created successfully", todo))
}

func GetTodos(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Username not found in context"))
		return
	}
	collections := db.GetDB().Collection(db.TodoCollection)
	ctx, cancel := CreateContext()
	defer cancel()
	cursor, err := collections.Find(ctx, bson.M{"username": username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to retrieve todos: "+err.Error()))
		return
	}
	var todos []models.Todo
	if err := cursor.All(ctx, &todos); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to decode todos: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse("Todos retrieved successfully", todos))
}

func DeleteTodo(c *gin.Context) {
	username := c.GetString(("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Username not found in context"))
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("ID parameter is required"))
		return
	}
	collections := db.GetDB().Collection((db.TodoCollection))
	ctx, cancel := CreateContext()
	defer cancel()
	// Delete the Todo with the specified ID and username
	result, err := collections.DeleteOne(ctx, bson.M{"id": id, "username": username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to delete todo: "+err.Error()))
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Todo not found"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse("Todo deleted successfully", nil))
}

func UpdateTodo(c *gin.Context) {
	// GEt the ID from the URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("ID parameter is required"))
		return
	}

	// Parse the JSON request body into the Todo struct
	var tmp struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindBodyWith(&tmp, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid JSON format: "+err.Error()))
		return
	}

	// Retrieve the username from the context
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Username not found in context"))
		return
	}

	// Create updated TODO object
	todo := bson.M{
		"title":       tmp.Title,
		"description": tmp.Description,
		"username":    username,
	}
	
	// Update the Todo in the database
	collections := db.GetDB().Collection(db.TodoCollection)
	ctx, cancel := CreateContext()
	defer cancel()

	filter := bson.M{"id": id, "username": username}
	update := bson.M{"$set": todo}
	// Update the document with the specified ID and username
	result, err := collections.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to update todo: "+err.Error()))
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Todo not found"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse("Todo updated successfully", nil))
}