// Images operations in db
// Functions needed:
//func CreateImage(ctx context.Context, image *model.Image) error
//func GetImageByID(ctx context.Context, imageID string, userID int64) (*model.Image, error)
//func GetImagesByUserID(ctx context.Context, userID int64, page, limit int) ([]model.Image, int, error)
//func UpdateImage(ctx context.Context, image *model.Image) error
//func DeleteImage(ctx context.Context, imageID string, userID int64) error

// Transformation functions:
//func CreateImageTransformation(ctx context.Context, transformation *model.ImageTransformation) error
//func GetTransformationsByImageID(ctx context.Context, imageID string) ([]model.ImageTransformation, error)
//func GetTransformationByParams(ctx context.Context, imageID string, params map[string]interface{}) (*model.ImageTransformation, error)

package db

import (
	"context"
	"database/sql"
	"fmt"
	"image-processing/model"
	"time"
)

func CreateImage(ctx context.Context, image *model.Image) error {
	query := `
		INSERT INTO images (id, user_id, original_filename, s3_key, s3_url, file_size, mime_type, created_at)
		VALUES(?,?,?,?,?,?,?,NOW())
	`
	_, err := db.ExecContext(ctx, query,
		image.ID,
		image.UserID,
		image.Filename,
		image.S3Key,
		image.URL,
		image.Size,
		image.MimeType,
	)
	if err != nil {
		return fmt.Errorf("failed to create image: %v", err)
	}
	return nil
}

func GetImageByID(ctx context.Context, imageID string, userID int64) (*model.Image, error) {
	var image model.Image
	var createdAtStr string // Use string intermediate for MySQL datetime

	query := `
		SELECT id, user_id, original_filename, s3_key, s3_url, file_size, mime_type, created_at
		FROM images WHERE id = ? AND user_id = ?
	`
	err := db.QueryRowContext(ctx, query, imageID, userID).Scan(
		&image.ID,
		&image.UserID,
		&image.Filename,
		&image.S3Key,
		&image.URL,
		&image.Size,
		&image.MimeType,
		&createdAtStr, // Scan as string first
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image not found")
		}
		return nil, fmt.Errorf("failed to get image details: %v", err)
	}

	// Parse the datetime string to time.Time
	parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	image.CreatedAt = parsedTime

	return &image, nil
}

func GetImagesByUserID(ctx context.Context, userID int64, page, limit int) ([]model.Image, int, error) {
	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Query to get images with pagination
	query := `
        SELECT id, user_id, original_filename, s3_key, s3_url, file_size, mime_type, created_at
        FROM images WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
    `

	// Use QueryContext for multiple rows, not QueryRowContext
	rows, err := db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []model.Image

	// Loop through all rows
	for rows.Next() {
		var image model.Image
		var createdAtStr string

		err := rows.Scan(
			&image.ID,
			&image.UserID,
			&image.Filename,
			&image.S3Key,
			&image.URL,
			&image.Size,
			&image.MimeType,
			&createdAtStr,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan image: %w", err)
		}

		// Parse the datetime string to time.Time
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse created_at: %w", err)
		}
		image.CreatedAt = parsedTime

		images = append(images, image)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM images WHERE user_id = ?`
	var totalCount int
	err = db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return images, totalCount, nil
}

func UpdateImage(ctx context.Context, image *model.Image) error {
	query := `
		UPDATE images
		SET original_filename = ?, s3_key = ?, s3_url = ?, file_size = ?, mime_type = ?
		WHERE id = ? AND user_id = ?
	`
	result, err := db.ExecContext(ctx, query,
		image.Filename,
		image.S3Key,
		image.URL,
		image.Size,
		image.MimeType,
		image.ID,
		image.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}
	// Check if any row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image not found or not owned by user")
	}
	return nil
}

func DeleteImage(ctx context.Context, imageID string, userID int64) error {
	query := `
        DELETE FROM images 
        WHERE id = ? AND user_id = ?
    `
	result, err := db.ExecContext(ctx, query, imageID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image not found or not owned by user")
	}

	return nil
}

// Transformation functions
func CreateImageTransformation(ctx context.Context, transformation *model.ImageTransformation) error {
	query := `
		INSERT INTO image_transformations (id, original_image_id, transformed_s3_key, transformed_s3_url, transformation_params, file_size, format, created_at)
		VALUES(?,?,?,?,?,?,?,NOW())
	`
	_, err := db.ExecContext(ctx, query,
		transformation.ID,
		transformation.OriginalImageID,
		transformation.S3Key,
		transformation.S3URL,
		transformation.Parameters,
		transformation.FileSize,
		transformation.Format,
	)
	if err != nil {
		return fmt.Errorf("failed to create image transformation: %v", err)
	}
	return nil
}

func GetTransformationsByImageID(ctx context.Context, imageID string) ([]model.ImageTransformation, error) {
	query := `
		SELECT id, original_image_id, transformed_s3_key, transformed_s3_url, transformation_params, file_size, format, created_at
		FROM image_transformations WHERE original_image_id = ? ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transformations: %w", err)
	}
	defer rows.Close()

	var transformations []model.ImageTransformation

	for rows.Next() {
		var transformation model.ImageTransformation
		var createdAtStr string

		err := rows.Scan(
			&transformation.ID,
			&transformation.OriginalImageID,
			&transformation.S3Key,
			&transformation.S3URL,
			&transformation.Parameters,
			&transformation.FileSize,
			&transformation.Format,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transformation: %w", err)
		}

		// Parse the datetime string to time.Time
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		transformation.CreatedAt = parsedTime

		transformations = append(transformations, transformation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return transformations, nil
}

func GetTransformationByID(ctx context.Context, transformationID string) (*model.ImageTransformation, error) {
	var transformation model.ImageTransformation
	var createdAtStr string

	query := `
		SELECT id, original_image_id, transformed_s3_key, transformed_s3_url, transformation_params, file_size, format, created_at
		FROM image_transformations WHERE id = ?
	`
	err := db.QueryRowContext(ctx, query, transformationID).Scan(
		&transformation.ID,
		&transformation.OriginalImageID,
		&transformation.S3Key,
		&transformation.S3URL,
		&transformation.Parameters,
		&transformation.FileSize,
		&transformation.Format,
		&createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transformation not found")
		}
		return nil, fmt.Errorf("failed to get transformation details: %v", err)
	}

	// Parse the datetime string to time.Time
	parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	transformation.CreatedAt = parsedTime

	return &transformation, nil
}

func GetTransformationByParams(ctx context.Context, imageID string, params string) (*model.ImageTransformation, error) {
	var transformation model.ImageTransformation
	var createdAtStr string

	query := `
		SELECT id, original_image_id, transformed_s3_key, transformed_s3_url, transformation_params, file_size, format, created_at
		FROM image_transformations WHERE original_image_id = ? AND transformation_params = ?
	`
	err := db.QueryRowContext(ctx, query, imageID, params).Scan(
		&transformation.ID,
		&transformation.OriginalImageID,
		&transformation.S3Key,
		&transformation.S3URL,
		&transformation.Parameters,
		&transformation.FileSize,
		&transformation.Format,
		&createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transformation not found")
		}
		return nil, fmt.Errorf("failed to get transformation details: %v", err)
	}

	// Parse the datetime string to time.Time
	parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	transformation.CreatedAt = parsedTime

	return &transformation, nil
}

func DeleteTransformation(ctx context.Context, transformationID string) error {
	query := `
        DELETE FROM image_transformations 
        WHERE id = ?
    `
	result, err := db.ExecContext(ctx, query, transformationID)
	if err != nil {
		return fmt.Errorf("failed to delete transformation: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transformation not found")
	}

	return nil
}
