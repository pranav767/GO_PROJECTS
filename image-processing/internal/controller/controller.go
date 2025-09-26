package controller

import (
	"image-processing/internal/db"
	"image-processing/internal/service"
	"image-processing/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ImageController struct {
	uploadService     *service.ImageUploadService
	processingService *service.ImageProcessingService
}

func NewImageController(uploadService *service.ImageUploadService, processingService *service.ImageProcessingService) *ImageController {
	return &ImageController{
		uploadService:     uploadService,
		processingService: processingService,
	}
}

func RegisterHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	err = service.RegisterUser(user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	exist, err := service.AuthenticateUser(user.Username, user.Password)
	if !exist {
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User does not exist"})
			case "invalid passwd":
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth failed"})
		}
		return
	}
	token, err := service.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ic *ImageController) UploadImage(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get file from multipart form
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Call the upload service with correct parameters
	err, resp := ic.uploadService.UploadImage(
		c.Request.Context(),
		file.Filename,
		file.Size,
		src, 	
		file.Header.Get("Content-Type"),
		uid,                             	
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (ic *ImageController) GetImages(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	// Get pagination parameters with defaults
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	// Get images from database
	images, total, err := db.GetImagesByUserID(c.Request.Context(), uid, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Create response
	response := &model.ListResponse{
		Images: images,
		Total:  total,
	}

	c.JSON(http.StatusOK, response)
}

func (ic *ImageController) GetImage(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get image ID from URL parameter
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	// Get image from database
	image, err := db.GetImageByID(c.Request.Context(), imageID, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, image)
}

func (ic *ImageController) DeleteImage(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get image ID from URL parameter
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	// Delete image using service
	err := ic.uploadService.DeleteImage(c.Request.Context(), imageID, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

func (ic *ImageController) TransformImage(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get image ID from URL parameter
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	// Parse transformation request from JSON body
	var req model.TransformationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transformation parameters", "details": err.Error()})
		return
	}

	// Apply transformations using service
	response, err := ic.processingService.TransformImage(c.Request.Context(), imageID, uid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ic *ImageController) GetTransformations(c *gin.Context) {
	// Extract user ID from JWT middleware
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get image ID from URL parameter
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	// Get all transformations for this image
	transformations, err := ic.processingService.GetTransformationsByImageID(c.Request.Context(), imageID, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_id":        imageID,
		"transformations": transformations,
		"count":           len(transformations),
	})
}
