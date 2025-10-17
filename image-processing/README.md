# Image Processing API Service

A comprehensive Go-based REST API service for image processing with JWT authentication, AWS S3 integration, and powerful transformation capabilities. Built with Gin framework and designed for production use.

> Project idea from: https://roadmap.sh/projects/image-processing-service

## 🌟 Features

- ✅ **JWT Authentication** - Secure user registration and login
- ✅ **Image Upload & Management** - Upload, list, retrieve, and delete images
- ✅ **AWS S3 Integration** - Scalable cloud storage with organized folder structure
- ✅ **Image Transformations** - Resize, rotate, flip, color adjustments, filters
- ✅ **Database Integration** - MySQL with comprehensive audit trails
- ✅ **Format Support** - JPEG, PNG with quality control
- ✅ **RESTful Design** - Clean, intuitive API endpoints
- 🚧 **Rate Limiting** - Coming soon for API protection
- 🚧 **Input Validation** - Enhanced validation for all endpoints

## 📋 Prerequisites

- **Go 1.21+**
- **MySQL 8.0+**
- **AWS Account** (for S3 storage)
- **Docker** (optional, for MySQL)

## 🛠️ Installation & Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd image-processing
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Configure Environment
Create `config.env` file:
```env
HMAC_SECRET=your-jwt-secret-key-here
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key-id
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
S3_BUCKET_NAME=your-s3-bucket-name
MAX_FILE_SIZE=10485760
ALLOWED_IMAGE_FORMATS=jpg,jpeg,png,gif,webp
```

### 4. Setup Database
```bash
# Using Docker
docker run --name mysql \
  -e MYSQL_ROOT_PASSWORD=adminpass \
  -e MYSQL_DATABASE=image_processing \
  -p 3306:3306 \
  -d mysql:8

# Apply database schema
sudo docker exec -i mysql mysql -u root -padminpass image_processing < internal/db/db.sql
```

### 5. Run Application
```bash
go run cmd/main.go
```
Server starts on `http://localhost:8080`

## 🔧 Project Structure

```
image-processing/
├── cmd/
│   └── main.go                          # Application entry point
├── internal/
│   ├── controller/
│   │   └── controller.go                # HTTP handlers & endpoints
│   ├── db/
│   │   ├── db.go                        # Database connection
│   │   ├── aws.go                       # AWS S3 client setup
│   │   ├── users.go                     # User operations
│   │   └── images.go                    # Image & transformation ops
│   ├── middleware/
│   │   └── jwt.go                       # JWT authentication
│   ├── routes/
│   │   └── route.go                     # Route definitions
│   └── service/
│       ├── auth.go                      # Authentication logic
│       ├── upload_image.go              # Image upload service
│       └── image_processing.go          # Transformation engine
├── model/
│   └── model.go                         # Data models & structs
├── utils/
│   └── utils.go                         # Utility functions
├── config.env                           # Environment configuration
├── test_simple_s3.go                    # S3 connection test
├── go.mod                               # Go module file
├── go.sum                               # Dependency checksums
└── README.md                            # This file
```

## 📡 API Endpoints

### Authentication Endpoints
| Method | Endpoint   | Description           | Auth Required |
|--------|------------|-----------------------|---------------|
| POST   | /register  | Create new user       | No            |
| POST   | /login     | Authenticate user     | No            |

### Image Management Endpoints
| Method | Endpoint                    | Description                    | Auth Required |
|--------|-----------------------------|--------------------------------|---------------|
| POST   | /images                     | Upload image                   | Yes           |
| GET    | /images                     | List user's images             | Yes           |
| GET    | /images/{id}               | Get specific image             | Yes           |
| DELETE | /images/{id}               | Delete image                   | Yes           |
| POST   | /images/{id}/transform     | Transform image                | Yes           |
| GET    | /images/{id}/transformations| List image transformations    | Yes           |

## 🚀 API Usage Examples

### Authentication Flow
```bash
# 1. Register new user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'

# 2. Login and get JWT token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'
# Save token from response for next steps

# 3. Upload image
curl -X POST http://localhost:8080/images \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "image=@/path/to/image.jpg"

# 4. Transform image
curl -X POST http://localhost:8080/images/IMAGE_ID/transform \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "width": 800,
    "height": 600,
    "resize": "fit",
    "quality": 95,
    "format": "jpeg"
  }'
```

## 🎨 Image Transformation Options

All transformation parameters are **optional** - include only what you need:

```json
{
  "width": 800,           // Resize width
  "height": 600,          // Resize height  
  "quality": 90,          // JPEG quality (1-100)
  "format": "jpeg",       // Output format: jpeg, png
  "resize": "fit",        // Resize mode: fit, fill, crop
  "rotate": 90,           // Rotation: 90, 180, 270 degrees
  "flip": "horizontal",   // Flip: horizontal, vertical
  "brightness": 15,       // Brightness: -100 to 100
  "contrast": 10,         // Contrast: -100 to 100
  "blur": 1.5,           // Blur radius
  "sharpen": 2.0,        // Sharpen amount
  "grayscale": true      // Convert to grayscale
}
```

### Resize Modes
- **`fit`** - Resize to fit within dimensions, maintaining aspect ratio
- **`fill`** - Resize to fill dimensions, may crop, maintaining aspect ratio  
- **`crop`** - Center crop to exact dimensions

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Images Table  
```sql
CREATE TABLE images (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    s3_url VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Image Transformations Table
```sql
CREATE TABLE image_transformations (
    id VARCHAR(36) PRIMARY KEY,
    original_image_id VARCHAR(36) NOT NULL,
    transformed_s3_key VARCHAR(500) NOT NULL,
    transformed_s3_url VARCHAR(1000) NOT NULL,
    transformation_params TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    format VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (original_image_id) REFERENCES images(id)
);
```

## 📦 S3 Storage Structure

```
your-s3-bucket/
├── users/
│   ├── 1/
│   │   ├── originals/
│   │   │   ├── uuid1.jpg
│   │   │   └── uuid2.png
│   │   └── transformations/
│   │       ├── transform-uuid1.jpg
│   │       └── transform-uuid2.jpg
│   └── 2/
│       ├── originals/
│       │   └── uuid3.jpg
│       └── transformations/
│           └── transform-uuid3.jpg
```

## 🔍 Testing & Debugging

### Test S3 Connection
```bash
go run test_simple_s3.go
```

### MySQL Connection Test
```bash
sudo docker exec -i mysql mysql -u root -padminpass -e "SHOW DATABASES;"
```

### Health Check Endpoints
```bash
# Test authentication
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test"}'

# Test image listing (with auth)
curl -X GET http://localhost:8080/images \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## ⚡ Performance & Scalability

### Current Implementation
- **Concurrent Processing** - Go goroutines for image transformations
- **Efficient Memory Usage** - Streaming uploads to S3
- **Database Connection Pooling** - MySQL connection management
- **Image Format Optimization** - JPEG quality control, format conversion

### Planned Optimizations
- **Caching Layer** - Redis for frequently accessed images
- **Background Processing** - Queue system for heavy transformations
- **CDN Integration** - CloudFront for global image delivery
- **Horizontal Scaling** - Load balancer support

## 🛡️ Security Features

### Current Security
- ✅ **JWT Authentication** - Secure token-based auth
- ✅ **Password Hashing** - bcrypt with salt
- ✅ **User Isolation** - Users can only access their own images
- ✅ **File Type Validation** - Only allowed image formats
- ✅ **File Size Limits** - Configurable upload limits
- ✅ **S3 Access Control** - AWS IAM permissions

## 📈 Future Development

### Phase 1: Enhanced Security & Performance
- [ ] **Rate Limiting** - Prevent API abuse
  - Request limits per user/IP
  - Sliding window implementation
  - Different limits for different endpoints
- [ ] **Input Validation** - Comprehensive request validation
  - Parameter sanitization
  - File content validation
  - Request size limits
- [ ] **API Documentation** - Swagger/OpenAPI integration
- [ ] **Metrics & Monitoring** - Prometheus integration

### Phase 2: Advanced Features
- [ ] **Batch Processing** - Multiple image operations
- [ ] **Watermarking** - Brand/copyright protection
- [ ] **Face Detection** - AI-powered image analysis
- [ ] **Format Conversion** - WebP, AVIF support
- [ ] **Image Optimization** - Automatic compression

## 🐛 Troubleshooting

### Common Issues

**Empty S3 Bucket**
- Check AWS credentials in `config.env`
- Verify S3 bucket exists and has proper permissions
- Review debug logs for upload errors

**Database Connection Errors**  
- Ensure MySQL container is running: `docker ps`
- Check database schema is applied
- Verify connection string in code

**JWT Token Issues**
- Ensure `HMAC_SECRET` is set in environment
- Check token format in Authorization header
- Verify token hasn't expired

**Image Upload Failures**
- Check file size limits (MAX_FILE_SIZE)
- Verify allowed file formats
- Ensure proper multipart form submission

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Development Guidelines
- Follow Go naming conventions
- Add tests for new features
- Update documentation
- Use meaningful commit messages

## � Project Inspiration

This project is based on the [Image Processing Service](https://roadmap.sh/projects/image-processing-service) project from **roadmap.sh** - a comprehensive backend development project designed to teach scalable image processing with user authentication and transformation capabilities.

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework
- [disintegration/imaging](https://github.com/disintegration/imaging) - Image processing library
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go-v2) - AWS services integration
- [golang-jwt](https://github.com/golang-jwt/jwt) - JWT implementation

**Built with ❤️ and Go**