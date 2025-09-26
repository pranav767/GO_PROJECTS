package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"image-processing/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func InitS3Connection() (*model.S3Client, error) {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error loading config.env file")
		return nil, err
	}
	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	accessID := os.Getenv("AWS_ACCESS_KEY_ID")
	accessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if s3Bucket == "" {
		return nil, errors.New("S3_BUCKET is required")
	}
	if awsRegion == "" {
		return nil, errors.New("AWS_REGION is required")
	}
	if accessID == "" {
		return nil, errors.New("ACCESS_ID is required")
	}
	if accessKey == "" {
		return nil, errors.New("ACCESS_KEY is required")
	}

	// Import S3 creds and load client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessID,
				SecretAccessKey: accessKey,
			},
		}))

	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	s3Client := &model.S3Client{
		Client: client,
		Bucket: s3Bucket,
		Region: awsRegion,
	}

	return s3Client, nil
}

func TestS3Connection(ctx context.Context, s3Client *model.S3Client) error {
	_, err := s3Client.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s3Client.Bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to access S3 bucket '%s': %w", s3Client.Bucket, err)
	}
	return nil
}
