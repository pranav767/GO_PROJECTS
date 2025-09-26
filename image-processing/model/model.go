package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// S3Client holds AWS S3 client configuration
type S3Client struct {
	Client *s3.Client
	Bucket string
	Region string
}

// Image represents the main image record
type Image struct {
	ID        string    `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Filename  string    `json:"filename" db:"original_filename"`
	S3Key     string    `json:"s3_key" db:"s3_key"`
	URL       string    `json:"url" db:"s3_url"`
	Size      int64     `json:"size" db:"file_size"`
	MimeType  string    `json:"mime_type" db:"mime_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ImageTransformation represents a transformed version
type ImageTransformation struct {
	ID              string    `json:"id" db:"id"`
	OriginalImageID string    `json:"original_image_id" db:"original_image_id"`
	S3Key           string    `json:"s3_key" db:"transformed_s3_key"`
	S3URL           string    `json:"s3_url" db:"transformed_s3_url"`
	Parameters      string    `json:"parameters" db:"transformation_params"`
	FileSize        int64     `json:"file_size" db:"file_size"`
	Format          string    `json:"format" db:"format"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Request structs
type TransformationRequest struct {
	Width      *int     `json:"width,omitempty"`      // Optional: resize width
	Height     *int     `json:"height,omitempty"`     // Optional: resize height
	Quality    *int     `json:"quality,omitempty"`    // Optional: 1-100 for JPEG quality
	Format     *string  `json:"format,omitempty"`     // Optional: jpeg, png, webp
	Resize     *string  `json:"resize,omitempty"`     // Optional: fit, fill, crop
	Rotate     *int     `json:"rotate,omitempty"`     // Optional: degrees: 90, 180, 270
	Flip       *string  `json:"flip,omitempty"`       // Optional: horizontal, vertical
	Brightness *int     `json:"brightness,omitempty"` // Optional: -100 to 100
	Contrast   *int     `json:"contrast,omitempty"`   // Optional: -100 to 100
	Blur       *float64 `json:"blur,omitempty"`       // Optional: blur radius
	Sharpen    *float64 `json:"sharpen,omitempty"`    // Optional: sharpen amount
	Grayscale  *bool    `json:"grayscale,omitempty"`  // Optional: convert to grayscale
}

// TransformationResponse contains the result of image transformation
type TransformationResponse struct {
	TransformationID string    `json:"transformation_id"`
	OriginalImageID  string    `json:"original_image_id"`
	S3URL            string    `json:"s3_url"`
	S3Key            string    `json:"s3_key"`
	Parameters       string    `json:"parameters"`
	FileSize         int64     `json:"file_size"`
	Format           string    `json:"format"`
	CreatedAt        time.Time `json:"created_at"`
}

// Legacy structs (kept for backward compatibility)
type TransformRequest struct {
	Resize *ResizeParams `json:"resize,omitempty"`
	Crop   *CropParams   `json:"crop,omitempty"`
	Rotate *int          `json:"rotate,omitempty"`
	Format *string       `json:"format,omitempty"`
}

type ResizeParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type CropParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// Response structs
type UploadResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type ListResponse struct {
	Images []Image `json:"images"`
	Total  int     `json:"total"`
}
