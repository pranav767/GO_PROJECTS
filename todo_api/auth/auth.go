package auth

import (
	"encoding/base64"
	"fmt"
	"todo_api/db"
	"todo_api/handler"
	"todo_api/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func Register(c *gin.Context) {
	// Bind the API info with body struct
	var body models.Users
	c.BindJSON(&body)

	// Check if user exists
	var user models.Users
	collection := db.GetDB().Collection("users")
	filter := bson.M{"email": body.Email}
	err := collection.FindOne(c, filter).Decode(&user)
	if err == nil {
		c.JSON(400, gin.H{"error": "EMail already registered"})
		return
	}

	hashedPassword := GenerateBase64Token(body.Password)
	// Add the hashed token to Users collection
	// Create user body
	newUser := models.Users{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hashedPassword),
	}
	ctx, cancel := handler.CreateContext()
	defer cancel()
	result, err := collection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(500, gin.H{"Error": "Failed to create user"})
	}
	c.JSON(200, gin.H{"message": "User registered Successfully", "id": result.InsertedID, "token": string(hashedPassword)})
}

func Login(c *gin.Context) {
	// Validate api inputs with body struct
	var body models.Users
	c.BindJSON(&body)

	// create a tmp body/struct for existing data as users struct.
	var user models.Users
	collection := db.GetDB().Collection("users")
	// Create a filter with the given username, and search in db.
	filter := bson.M{"username": body.Username}
	err := collection.FindOne(c, filter).Decode(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "No Such User Registered"})
		return
	}
	// Create base64 token for the entered password and compare with real password.
	hashedPassword := GenerateBase64Token(body.Password)
	fmt.Printf(hashedPassword, user.Password)
	if hashedPassword == user.Password {
		c.JSON(200, gin.H{"success": "Authentication Success full", "token": hashedPassword})
		return
	} else {
		c.JSON(403, gin.H{"Failed": "Incorrect Password"})
		return
	}

}

// Create a basic base64 token
func GenerateBase64Token(password string) string {
	// Hash the password
	return base64.StdEncoding.EncodeToString([]byte(password))
}

func DecodeToken(token string) (string, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	return string(decodedToken), nil // Return email or identifier
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Decode the token to retrieve the password
		decodedPassword, err := DecodeToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		var body struct {
			Username string `json:"username"`
		}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		// Get the username from the request parameter
		username := body.Username
		if username == "" {
			c.JSON(400, gin.H{"error": "Username parameter is required"})
			c.Abort()
			return
		}

		// Retrieve user document from the database using the username
		collection := db.GetDB().Collection("users")
		var user models.Users
		filter := bson.M{"username": username}
		ctx, cancel := handler.CreateContext()
		defer cancel()
		err = collection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			c.JSON(404, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		realdecodedPassword, err := DecodeToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Error while decoding token"})
			c.Abort()
			return
		}
		// Compare the decoded token (password) with the stored password
		if decodedPassword != realdecodedPassword {
			c.JSON(403, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user document in context for downstream handlers
		c.Set("username", string(username))
		c.Next()
	}
}
