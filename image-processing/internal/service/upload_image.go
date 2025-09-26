package service

import (
	"context"
	"errors"
	"fmt"
	"image-processing/internal/db"
	"image-processing/model"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ImageUploadService struct {
	s3Client *model.S3Client
}

func NewImageUploadService(s3client *model.S3Client) *ImageUploadService {
	return &ImageUploadService{s3Client: s3client}
}

func (s *ImageUploadService) UploadImage(ctx context.Context, filename string, fileSize int64, file io.Reader, contentType string, userID int64) (*model.UploadResponse, error) {
	// 1. Validate the image
	if err := s.validateFile(fileSize, contentType); err != nil {
		return nil, err
	}
	//2. Generate a UUID
	imageID := uuid.New().String()

	// 3.Gen s3 key
	s3Key := fmt.Sprintf("users/%d/original/%s%s", userID, imageID, filepath.Ext(filename))
	// 4.Upload to s3
	s3URl, err := s.uploadToS3(ctx, s3Key, file, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}
	// 5. Create db record
	image := &model.Image{
		ID:        imageID,
		UserID:    userID,
		Filename:  filename,
		S3Key:     s3Key,
		URL:       s3URl,
		Size:      fileSize,
		MimeType:  contentType,
		CreatedAt: time.Now(),
	}
	// 6.update db
	if err := db.CreateImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to save image in db: %w", err)
	}
	// 7. return response.
	response := &model.UploadResponse{
		ID:       imageID,
		URL:      s3URl,
		Filename: filename,
	}
	return response, nil

}

func (s *ImageUploadService) validateFile(fileSize int64, contentType string) error {
	//1. CheckfileSize
	//2 .Checkfiletype
	maxSize := int64(10 << 20) // 10 MB Size
	if fileSize > maxSize {
		return errors.New("file size greater than 10MB")
	}
	fileType := []string{"image/jpg", "image/png", "image/jpeg", "image/gif"}
	for _, allowed := range fileType {
		if allowed == contentType {
			return nil
		}
	}
	return fmt.Errorf("unSuppported file type %s", contentType)
}

func (s *ImageUploadService) uploadToS3(ctx context.Context, s3Key string, file io.Reader, contentType string) (string, error) {
	// https://pkg.go.dev/github.com/aws/aws-sdk-go/service/s3#S3.PutObject

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.s3Client.Bucket),
		Key:         &s3Key,
		Body:        file,
		ContentType: &contentType,
	}
	_, err := s.s3Client.Client.PutObject(ctx, input)
	if err != nil {
		return "", err
	}
	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.s3Client.Bucket, s.s3Client.Region, s3Key)
	return s3URL, nil
}

func (s *ImageUploadService) DeleteImage(ctx context.Context, imageID string, userID int64) error {
	// 1. First, get the image details to retrieve S3 key (and verify ownership)
	image, err := db.GetImageByID(ctx, imageID, userID)
	if err != nil {
		return fmt.Errorf("failed to find image: %w", err)
	}

	// 2. Delete from database first (ensures consistency)
	err = db.DeleteImage(ctx, imageID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}

	// 3. Delete from S3 (if this fails, log error but don't fail the operation)
	err = s.deleteFromS3(ctx, image.S3Key)
	if err != nil {
		// Log the error but don't fail the deletion since DB record is already gone
		// In production, you might want to queue this for retry
		fmt.Printf("Warning: Failed to delete image from S3 (key: %s): %v\n", image.S3Key, err)
		return fmt.Errorf("image deleted from database but S3 cleanup failed: %w", err)
	}

	return nil
}

func (s *ImageUploadService) deleteFromS3(ctx context.Context, s3Key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.s3Client.Bucket),
		Key:    aws.String(s3Key),
	}

	_, err := s.s3Client.Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
