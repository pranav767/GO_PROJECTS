# Blogging Platform API

A RESTful API for managing blog posts built with Go, Gin, and MongoDB.

Based on: [roadmap.sh/projects/blogging-platform-api](https://roadmap.sh/projects/blogging-platform-api)

## Features

- CRUD operations for blog posts
- MongoDB integration for data persistence
- Auto-incrementing numeric IDs with thread safety
- Structured API responses with consistent formatting
- Clean code architecture with separation of concerns
- Docker support for MongoDB

## Project Structure

```
blogging_platform/
├── db/
│   └── mongo.go       # MongoDB connection and operations
├── handler/
│   ├── handler.go     # API endpoint handlers
│   └── utils.go       # Utility functions for handlers
├── models/
│   └── blog_post.go   # Data models and response structures
├── main.go            # Application entry point
└── README.md          # Project documentation
```

## Prerequisites

- Go 1.19 or higher
- Docker and Docker Compose (for MongoDB)

## Getting Started

### 1. Start MongoDB

Run MongoDB as a Docker container:

```bash
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=adminpass \
  mongo:latest
```

### 2. Run the API

```bash
# Navigate to the project directory
cd blogging_platform

# Install dependencies
go mod tidy

# Run the application
go run main.go
```

The server will start on http://localhost:8080

## API Endpoints

| Method | Endpoint    | Description              |
|--------|-------------|--------------------------|
| GET    | /ping       | Health check             |
| POST   | /posts      | Create a new blog post   |
| GET    | /posts      | Get all blog posts       |
| PUT    | /posts/:id  | Update a blog post by ID |
| DELETE | /posts/:id  | Delete a blog post by ID |

## Example Usage

### Health Check

```bash
curl http://localhost:8080/ping
```

Response:
```json
{
  "message": "pong"
}
```

### Create a Blog Post

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Blog Post",
    "content": "This is the content of my first blog post.",
    "category": "Technology",
    "tags": ["Tech", "Programming"]
  }'
```

Response:
```json
{
  "success": true,
  "message": "Blog post created successfully",
  "data": {
    "id": 1,
    "mongo_id": "64a1f3d288e4646b1f88bdd2",
    "post": {
      "id": "1",
      "numeric_id": 1,
      "title": "My First Blog Post",
      "content": "This is the content of my first blog post.",
      "category": "Technology",
      "tags": ["Tech", "Programming"],
      "created_at": "2025-07-24T15:04:10.684208535+05:30"
    }
  }
}
```

### Get All Blog Posts

```bash
curl http://localhost:8080/posts
```

Response:
```json
{
  "success": true,
  "message": "Blog posts retrieved successfully",
  "data": {
    "count": 1,
    "posts": [
      {
        "id": "1",
        "numeric_id": 1,
        "title": "My First Blog Post",
        "content": "This is the content of my first blog post.",
        "category": "Technology",
        "tags": ["Tech", "Programming"],
        "created_at": "2025-07-24T09:34:10.684Z"
      }
    ]
  }
}
```

### Update a Blog Post

```bash
curl -X PUT http://localhost:8080/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "content": "Updated content",
    "category": "Updated Category",
    "tags": ["updated", "new"]
  }'
```

Response:
```json
{
  "success": true,
  "message": "Blog post updated successfully",
  "data": {
    "id": 1,
    "post": {
      "id": "1",
      "numeric_id": 1,
      "title": "Updated Title",
      "content": "Updated content",
      "category": "Updated Category",
      "tags": ["updated", "new"],
      "updated_at": "2025-07-24T15:10:22.123456789+05:30"
    }
  }
}
```

### Delete a Blog Post

```bash
curl -X DELETE http://localhost:8080/posts/1
```

Response:
```json
{
  "success": true,
  "message": "Blog post deleted successfully",
  "data": {
    "id": "1"
  }
}
```

## Data Models

### Blog Post

```go
// BlogPost represents a blog post entity
type BlogPost struct {
    ID        string    `json:"id" bson:"id_str,omitempty"`
    NumericID int       `json:"numeric_id" bson:"id"`
    Title     string    `json:"title" binding:"required" bson:"title"`
    Content   string    `json:"content" binding:"required" bson:"content"`
    Category  string    `json:"category" binding:"required" bson:"category"`
    Tags      []string  `json:"tags" binding:"required" bson:"tags"`
    CreatedAt time.Time `json:"created_at,omitempty" bson:"createdAt"`
    UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updatedAt,omitempty"`
}
```

### API Response

```go
// Response represents a standard API response
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Count   int         `json:"count,omitempty"`
}
```

## Implementation Details

### Auto-incrementing IDs

The system uses a simple yet effective method to generate unique numeric IDs for blog posts:

1. When creating a new post, it queries the database for the document with the highest ID
2. Increments that value by 1 to get the next available ID
3. Uses a mutex to prevent race conditions when multiple requests arrive simultaneously

```go
// GetNextID returns a unique incremental ID
func GetNextID() (int, error) {
    // Use mutex to prevent race conditions
    mutex.Lock()
    defer mutex.Unlock()

    // Find the document with the highest ID
    opts := options.FindOne().SetSort(bson.M{"id": -1})
    var result bson.M

    err := collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
    if err != nil {
        // If no documents exist yet, start with ID 1
        return 1, nil
    }

    // Get the highest ID and increment by 1
    highestID := getIntFromBSON(result["id"])
    return highestID + 1, nil
}
```

### MongoDB Connection

The application maintains a singleton MongoDB client instance that's initialized at startup and reused across all requests:

```go
// DbInit initializes the MongoDB connection
func DbInit() error {
    var initErr error
    // Use sync.Once to ensure this only runs once
    once.Do(func() {
        initErr = initializeClient()
    })
    return initErr
}
```

### Error Handling

All API endpoints have consistent error handling with appropriate HTTP status codes and descriptive error messages.

## Monitoring MongoDB

You can interact with your MongoDB instance using the mongo shell:

```bash
# Connect to MongoDB with authentication
docker exec -it mongodb mongosh -u admin -p adminpass

# Select database
use blog_db

# List collections
show collections

# Query documents
db.blog_posts.find()
```

### MongoDB Setup with Docker Compose

For a more persistent setup, you can use Docker Compose:

Create a `docker-compose.yml` file:

```yaml
version: '3'

services:
  mongodb:
    image: mongo:latest
    container_name: mongodb
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: adminpass
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:
```

Then start with:

```bash
docker-compose up -d
```

## Shutdown

To stop the MongoDB container:

```bash
docker stop mongodb
docker rm mongodb
```

Or if using Docker Compose:

```bash
docker-compose down
```

## Future Improvements

- Add pagination for the GET /posts endpoint
- Add user authentication and authorization
- Implement search functionality by title/content
- Add categories and tags filtering
- Add unit and integration tests
- Add API documentation with Swagger/OpenAPI

## License

This project is licensed under the MIT License.
