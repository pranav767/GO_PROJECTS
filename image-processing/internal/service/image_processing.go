package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image-processing/internal/db"
	"image-processing/model"
	"image/jpeg"
	"image/png"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// ImageProcessingService handles all image transformation operations
type ImageProcessingService struct {
	s3Client *model.S3Client
}

// NewImageProcessingService creates a new image processing service
func NewImageProcessingService(s3Client *model.S3Client) *ImageProcessingService {
	return &ImageProcessingService{s3Client: s3Client}
}

// TransformImage applies the requested transformations to an image
func (ips *ImageProcessingService) TransformImage(ctx context.Context, imageID string, userID int64, req *model.TransformationRequest) (*model.TransformationResponse, error) {
	// 1. Get original image metadata and verify ownership
	originalImage, err := db.GetImageByID(ctx, imageID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original image: %w", err)
	}

	// 2. Download original image from S3
	originalImageData, err := ips.downloadFromS3(ctx, originalImage.S3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download image from S3: %w", err)
	}

	// 3. Decode the image
	img, originalFormat, err := image.Decode(bytes.NewReader(originalImageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 4. Apply transformations
	transformedImg := ips.applyTransformations(img, req)

	// 5. Determine output format
	outputFormat := originalFormat // Default to original format
	if req.Format != nil {
		outputFormat = *req.Format
	}

	// 6. Determine quality
	quality := 85 // Default quality
	if req.Quality != nil {
		quality = *req.Quality
	}

	// 7. Encode the transformed image
	transformedData, err := ips.encodeImage(transformedImg, outputFormat, quality)
	if err != nil {
		return nil, fmt.Errorf("failed to encode transformed image: %w", err)
	}

	// 8. Generate transformation ID and S3 key
	transformationID := uuid.New().String()
	s3Key := fmt.Sprintf("users/%d/transformations/%s.%s",
		userID, transformationID, outputFormat)

	// 9. Upload transformed image to S3
	s3URL, err := ips.uploadToS3(ctx, s3Key, bytes.NewReader(transformedData),
		fmt.Sprintf("image/%s", outputFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to upload transformed image to S3: %w", err)
	}

	// 10. Save transformation record to database
	transformation := &model.ImageTransformation{
		ID:              transformationID,
		OriginalImageID: imageID,
		S3Key:           s3Key,
		S3URL:           s3URL,
		Parameters:      ips.serializeParameters(req),
		FileSize:        int64(len(transformedData)),
		Format:          outputFormat,
		CreatedAt:       time.Now(),
	}

	err = db.CreateImageTransformation(ctx, transformation)
	if err != nil {
		return nil, fmt.Errorf("failed to save transformation record: %w", err)
	}

	// 11. Return response
	return &model.TransformationResponse{
		TransformationID: transformationID,
		OriginalImageID:  imageID,
		S3URL:            s3URL,
		S3Key:            s3Key,
		Parameters:       transformation.Parameters,
		FileSize:         transformation.FileSize,
		Format:           outputFormat,
		CreatedAt:        transformation.CreatedAt,
	}, nil
}

// applyTransformations applies all requested transformations to the image
func (ips *ImageProcessingService) applyTransformations(img image.Image, req *model.TransformationRequest) image.Image {
	result := img

	// Apply resize transformations first
	if req.Width != nil && *req.Width > 0 || req.Height != nil && *req.Height > 0 {
		result = ips.applyResize(result, req)
	}

	// Apply rotation
	if req.Rotate != nil && *req.Rotate != 0 {
		result = ips.applyRotation(result, *req.Rotate)
	}

	// Apply flip
	if req.Flip != nil && *req.Flip != "" {
		result = ips.applyFlip(result, *req.Flip)
	}

	// Apply color adjustments
	if req.Brightness != nil && *req.Brightness != 0 {
		result = imaging.AdjustBrightness(result, float64(*req.Brightness))
	}

	if req.Contrast != nil && *req.Contrast != 0 {
		result = imaging.AdjustContrast(result, float64(*req.Contrast))
	}

	// Apply filters
	if req.Blur != nil && *req.Blur > 0 {
		result = imaging.Blur(result, *req.Blur)
	}

	if req.Sharpen != nil && *req.Sharpen > 0 {
		result = imaging.Sharpen(result, *req.Sharpen)
	}

	if req.Grayscale != nil && *req.Grayscale {
		result = imaging.Grayscale(result)
	}

	return result
}

// applyResize handles different resize modes
func (ips *ImageProcessingService) applyResize(img image.Image, req *model.TransformationRequest) image.Image {
	width := 0
	height := 0

	if req.Width != nil {
		width = *req.Width
	}
	if req.Height != nil {
		height = *req.Height
	}

	// Default resize mode is "fit"
	resizeMode := "fit"
	if req.Resize != nil {
		resizeMode = *req.Resize
	}

	switch strings.ToLower(resizeMode) {
	case "fit":
		// Resize to fit within dimensions, maintaining aspect ratio
		return imaging.Fit(img, width, height, imaging.Lanczos)
	case "fill":
		// Resize to fill dimensions, may crop, maintaining aspect ratio
		return imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
	case "crop":
		// Center crop to exact dimensions
		return imaging.CropAnchor(img, width, height, imaging.Center)
	default:
		// Default resize behavior - maintain aspect ratio
		if width > 0 && height > 0 {
			return imaging.Fit(img, width, height, imaging.Lanczos)
		} else if width > 0 {
			return imaging.Resize(img, width, 0, imaging.Lanczos)
		} else if height > 0 {
			return imaging.Resize(img, 0, height, imaging.Lanczos)
		}
	}
	return img
}

// applyRotation handles image rotation
func (ips *ImageProcessingService) applyRotation(img image.Image, degrees int) image.Image {
	switch degrees {
	case 90, -270:
		return imaging.Rotate90(img)
	case 180, -180:
		return imaging.Rotate180(img)
	case 270, -90:
		return imaging.Rotate270(img)
	default:
		// For arbitrary angles, use the general rotate function
		return imaging.Rotate(img, float64(degrees), image.Transparent)
	}
}

// applyFlip handles image flipping
func (ips *ImageProcessingService) applyFlip(img image.Image, direction string) image.Image {
	switch strings.ToLower(direction) {
	case "horizontal":
		return imaging.FlipH(img)
	case "vertical":
		return imaging.FlipV(img)
	default:
		return img
	}
}

// encodeImage encodes the image to the specified format
func (ips *ImageProcessingService) encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	// Set default quality if not specified
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		return buf.Bytes(), err
	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	default:
		// Default to JPEG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		return buf.Bytes(), err
	}
}

// downloadFromS3 downloads an image from S3
func (ips *ImageProcessingService) downloadFromS3(ctx context.Context, s3Key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(ips.s3Client.Bucket),
		Key:    aws.String(s3Key),
	}

	result, err := ips.s3Client.Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object data: %w", err)
	}

	return data, nil
}

// uploadToS3 uploads the transformed image to S3
func (ips *ImageProcessingService) uploadToS3(ctx context.Context, s3Key string, data io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(ips.s3Client.Bucket),
		Key:         aws.String(s3Key),
		Body:        data,
		ContentType: aws.String(contentType),
	}

	_, err := ips.s3Client.Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		ips.s3Client.Bucket, ips.s3Client.Region, s3Key)

	return s3URL, nil
}

// serializeParameters converts transformation parameters to a JSON string for database storage
func (ips *ImageProcessingService) serializeParameters(req *model.TransformationRequest) string {
	params := make([]string, 0)

	if req.Width != nil && *req.Width > 0 {
		params = append(params, fmt.Sprintf("width=%d", *req.Width))
	}
	if req.Height != nil && *req.Height > 0 {
		params = append(params, fmt.Sprintf("height=%d", *req.Height))
	}
	if req.Quality != nil && *req.Quality > 0 {
		params = append(params, fmt.Sprintf("quality=%d", *req.Quality))
	}
	if req.Format != nil && *req.Format != "" {
		params = append(params, fmt.Sprintf("format=%s", *req.Format))
	}
	if req.Resize != nil && *req.Resize != "" {
		params = append(params, fmt.Sprintf("resize=%s", *req.Resize))
	}
	if req.Rotate != nil && *req.Rotate != 0 {
		params = append(params, fmt.Sprintf("rotate=%d", *req.Rotate))
	}
	if req.Flip != nil && *req.Flip != "" {
		params = append(params, fmt.Sprintf("flip=%s", *req.Flip))
	}
	if req.Brightness != nil && *req.Brightness != 0 {
		params = append(params, fmt.Sprintf("brightness=%d", *req.Brightness))
	}
	if req.Contrast != nil && *req.Contrast != 0 {
		params = append(params, fmt.Sprintf("contrast=%d", *req.Contrast))
	}
	if req.Blur != nil && *req.Blur > 0 {
		params = append(params, fmt.Sprintf("blur=%s", strconv.FormatFloat(*req.Blur, 'f', 2, 64)))
	}
	if req.Sharpen != nil && *req.Sharpen > 0 {
		params = append(params, fmt.Sprintf("sharpen=%s", strconv.FormatFloat(*req.Sharpen, 'f', 2, 64)))
	}
	if req.Grayscale != nil && *req.Grayscale {
		params = append(params, "grayscale=true")
	}

	return strings.Join(params, "&")
}

// GetTransformationsByImageID retrieves all transformations for a specific image
func (ips *ImageProcessingService) GetTransformationsByImageID(ctx context.Context, imageID string, userID int64) ([]model.ImageTransformation, error) {
	// First verify the user owns the original image
	_, err := db.GetImageByID(ctx, imageID, userID)
	if err != nil {
		return nil, fmt.Errorf("image not found or access denied: %w", err)
	}

	// Get all transformations for this image
	transformations, err := db.GetTransformationsByImageID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transformations: %w", err)
	}

	return transformations, nil
}

// DeleteTransformation removes a transformation (both from DB and S3)
func (ips *ImageProcessingService) DeleteTransformation(ctx context.Context, transformationID string, userID int64) error {
	// Get transformation details
	transformation, err := db.GetTransformationByID(ctx, transformationID)
	if err != nil {
		return fmt.Errorf("transformation not found: %w", err)
	}

	// Verify user owns the original image
	_, err = db.GetImageByID(ctx, transformation.OriginalImageID, userID)
	if err != nil {
		return fmt.Errorf("access denied: %w", err)
	}

	// Delete from database first
	err = db.DeleteTransformation(ctx, transformationID)
	if err != nil {
		return fmt.Errorf("failed to delete transformation from database: %w", err)
	}

	// Delete from S3 (log errors but don't fail)
	err = ips.deleteFromS3(ctx, transformation.S3Key)
	if err != nil {
		fmt.Printf("Warning: Failed to delete transformation from S3 (key: %s): %v\n", transformation.S3Key, err)
	}

	return nil
}

// deleteFromS3 deletes an object from S3
func (ips *ImageProcessingService) deleteFromS3(ctx context.Context, s3Key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(ips.s3Client.Bucket),
		Key:    aws.String(s3Key),
	}

	_, err := ips.s3Client.Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
