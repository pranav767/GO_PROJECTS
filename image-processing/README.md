# Image Processing API

A Go-based REST API for image processing with JWT authentication built using Gin framework.

## 🚀 Current Status

Currently implemented:
- ✅ JWT Authentication (Register/Login)
- ✅ User Management
- ✅ MySQL Database Integration
- 🚧 Image Processing (Coming Soon)

## 📋 Prerequisites

- Go 1.21+
- MySQL 8.0
- Docker (optional)

## 🛠 Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd image-processing
```

### 2. Environment Setup
Create a `.env` file in the root directory:
```bash
HMAC_SECRET=your_super_secret_jwt_key_here
```

### 3. Database Setup (Docker)
```bash
# Run MySQL container
docker run --name mysql-image \
  -e MYSQL_ROOT_PASSWORD=adminpass \
  -e MYSQL_DATABASE=image-processing \
  -p 3306:3306 \
  -d mysql:8

# Connect to MySQL (optional)
docker exec -it mysql-image mysql -u root -p
```

### 4. Install Dependencies
```bash
go mod tidy
```

### 5. Run the Application
```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080`

## 🔐 Authentication API

### Register User
Creates a new user account.

**Endpoint:** `POST /signup`

**Request Body:**
```json
{
  "username": "admin",
  "password": "adminpass"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "adminpass"
  }'
```

**Success Response:**
```json
{
  "message": "User registered successfully"
}
```

**Error Responses:**
```json
{
  "error": "user already exists"
}
```

---

### Login User
Authenticates a user and returns a JWT token.

**Endpoint:** `POST /login`

**Request Body:**
```json
{
  "username": "admin",
  "password": "adminpass"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "adminpass"
  }'
```

**Success Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
```json
{
  "error": "User does not exist"
}
```
```json
{
  "error": "Invalid password"
}
```

---

## 🧪 Testing the API

### Complete Authentication Flow

1. **Register a new user:**
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

2. **Login and get token:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

3. **Save the token from response and use it for protected routes:**
```bash
# Example for future protected endpoints
TOKEN="your_jwt_token_here"

curl -X GET http://localhost:8080/protected-endpoint \
  -H "Authorization: Bearer $TOKEN"
```

### Test with Multiple Users
```bash
# User 1
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "alice123"}'

# User 2  
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username": "bob", "password": "bob456"}'

# Login as Alice
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "alice123"}'
```

## 📁 Project Structure

```
image-processing/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── controller/
│   │   └── controller.go       # HTTP handlers
│   ├── db/
│   │   ├── db.go              # Database connection
│   │   ├── users.go           # User database operations
│   │   └── aws.go             # AWS integration
│   ├── middleware/
│   │   └── jwt.go             # JWT middleware
│   └── service/
│       └── auth.go            # Authentication business logic
├── model/
│   └── model.go               # Data models
├── utils/
│   └── utils.go               # Utility functions
├── go.mod
├── go.sum
└── README.md
```

## 🔧 Configuration

### Environment Variables
- `HMAC_SECRET`: Secret key for JWT signing (required)

### Database Configuration
Update the DSN in `internal/db/db.go`:
```go
const DSN = "root:adminpass@tcp(localhost:3306)/image-processing"
```

## 🐛 Troubleshooting

### Database Connection Issues
If you're getting "nil pointer dereference" errors:

1. **Ensure MySQL container is running:**
```bash
docker ps
# Should show mysql-image container running
```

2. **Check if database exists:**
```bash
docker exec -it mysql-image mysql -u root -p
# Enter password: adminpass
# Then run: SHOW DATABASES;
```

3. **Create database if it doesn't exist:**
```bash
docker exec -it mysql-image mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS \`image-processing\`;"
```

4. **Create users table:**
```sql
USE `image-processing`;
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

5. **Test database connection:**
```bash
# Check if you can connect with Go's DSN format
docker exec -it mysql-image mysql -u root -p -h localhost -P 3306 image-processing
```

### Common Errors and Solutions

**Error: "404 /register"** 
- Solution: Use `/signup` endpoint instead of `/register`

**Error: "nil pointer dereference"**
- Solution: Database connection failed. Check MySQL container and DSN configuration

**Error: "Invalid JWT token"**
- Solution: Make sure `HMAC_SECRET` is set in your `.env` file

## 🚧 Roadmap

- [ ] Image upload endpoints
- [ ] Image processing operations (resize, crop, filter)
- [ ] File storage (local/AWS S3)
- [ ] Rate limiting
- [ ] API documentation (Swagger)
- [ ] Docker compose setup
- [ ] Unit tests

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

---

**Note:** This API is currently in development. Image processing features will be added in upcoming releases.