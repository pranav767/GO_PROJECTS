# URL Shortener Service

A simple URL shortener built with Go, Gin, and MongoDB.

## Features

- Shorten long URLs to short codes
- Redirect short codes to original URLs
- Update and delete short codes
- Track access count (stats)
- Get all details for a short code
- RESTful API with JSON responses

## Project Structure

```
cmd/main.go                  # Application entry point
internal/controllers/        # HTTP handlers/controllers
internal/services/           # Business logic and MongoDB operations
internal/repository/mongo.go # MongoDB connection and helpers
internal/models/models.go    # Data models
```

## Setup

1. **MongoDB:**  
   Make sure MongoDB is running locally (default: `mongodb://localhost:27017`).

2. **Install dependencies:**  
   ```bash
   go mod tidy
   ```

3. **Run the server:**  
   ```bash
   go run cmd/main.go
   ```

## API Endpoints & Example `curl` Commands

### 1. Shorten a URL

**Request:**
```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"url":"https://example.com"}' \
     http://localhost:8080/shorten
```

**Response:**
```json
{
  "id": 1,
  "url": "https://example.com",
  "shortCode": "abc12345",
  "createdAt": "...",
  "updatedAt": "...",
  "accessCount": 0
}
```

---

### 2. Redirect to Original URL

**Request:**
```bash
curl -v http://localhost:8080/r/{shortCode}
```
or visit in your browser:  
`http://localhost:8080/r/{shortCode}`

---

### 3. Update a Short URL

**Request:**
```bash
curl -X PUT -H "Content-Type: application/json" \
     -d '{"url":"https://newexample.com"}' \
     http://localhost:8080/update/{shortCode}
```

---

### 4. Delete a Short URL

**Request:**
```bash
curl -X DELETE http://localhost:8080/delete/{shortCode}
```

---

### 5. Get Stats (Access Count)

**Request:**
```bash
curl http://localhost:8080/stats/{shortCode}
```

**Response:**
```json
{
  "shortCode": "abc12345",
  "accessCount": 5
}
```

---

### 6. Get All Details for a Short Code

**Request:**
```bash
curl http://localhost:8080/details/{shortCode}
```

**Response:**
```json
{
  "id": 1,
  "url": "https://example.com",
  "shortCode": "abc12345",
  "createdAt": "...",
  "updatedAt": "...",
  "accessCount": 5
}
```

---

## Notes

- Replace `{shortCode}` with the actual code returned from the shorten endpoint.
- All endpoints return JSON responses.
- Access count is incremented every time a redirect is performed.