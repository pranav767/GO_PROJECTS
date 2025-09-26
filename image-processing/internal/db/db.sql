CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Main Image table (simplified to match model.go)
CREATE TABLE IF NOT EXISTS images(
    id VARCHAR(36) PRIMARY KEY, -- UUID FOR IMAGES
    user_id BIGINT NOT NULL, 
    original_filename VARCHAR(255) NOT NULL, 
    s3_key VARCHAR(500) NOT NULL, 
    s3_url VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- transformed image table (simplified to match model.go)
CREATE TABLE IF NOT EXISTS image_transformations(
    id VARCHAR(36) PRIMARY KEY, -- UUID FOR TRANSFORMATIONS
    original_image_id VARCHAR(36) NOT NULL, 
    transformed_s3_key VARCHAR(500) NOT NULL,
    transformed_s3_url VARCHAR(1000) NOT NULL,
    transformation_params TEXT NOT NULL, -- Store as query string format
    file_size BIGINT NOT NULL,
    format VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (original_image_id) REFERENCES images(id) ON DELETE CASCADE,
    INDEX idx_original_image_id (original_image_id),
    INDEX idx_created_at (created_at)
);