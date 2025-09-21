# E-Commerce Go API (WIP)

This project is a simple e-commerce backend API written in Go, using Gin, MySQL, and JWT authentication. It demonstrates user registration, login, JWT-protected routes, and best practices for structuring a Go web project.

## Features
- User registration (`/signup`)
- User login with JWT issuance (`/login`)
- Passwords are securely hashed
- JWT authentication and middleware for protected routes
- Example protected route: `/profile` (returns username from JWT)
- Modular project structure (controllers, service, db, middleware, routes)

## Project Structure
```
cmd/main.go                # Entry point, initializes DB and server
internal/controller/       # HTTP handlers
internal/service/          # Business logic
internal/db/               # Database connection and queries
internal/routes/           # Route registration
internal/middleware/       # JWT middleware
utils/                     # Utility functions (hashing, JWT)
```

## Setup
1. **Start MySQL with Docker:**
   ```sh
   docker run -d \
     --name mysql \
     -p 3306:3306 \
     -e MYSQL_ROOT_PASSWORD=adminpass \
     -e MYSQL_DATABASE=e-commerce \
     -e MYSQL_USER=admin \
     -e MYSQL_PASSWORD=adminpass \
     mysql:latest
   ```
2. **Create the users table:**
   ```sh
   docker exec -it mysql bash
   mysql -u root -p
   # Enter password: adminpass
   USE `e-commerce`;
   CREATE TABLE users (
     id INT AUTO_INCREMENT PRIMARY KEY,
     username VARCHAR(255) NOT NULL UNIQUE,
     password_hash VARCHAR(255) NOT NULL
   );
   exit
   exit
   ```
3. **Set environment variables:**
   - Create a `.env` file in the project root:
     ```
     HMAC_SECRET=your-very-secret-key
     ```

4. **Run the Go server:**
   ```sh
   go run ./cmd/main.go
   ```

## API Usage Examples

### Register (Sign Up)
```sh
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"testpass"}'
```

### Login (Get JWT)
```sh
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"testpass"}'
```
- Response: `{ "token": "<jwt-token>" }`

### Access Protected Route
```sh
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer <jwt-token>"
```
- Response: `{ "username": "testuser" }`

## TODO / In Progress
- Input validation improvements
- Enhanced error handling and logging
- Move user model and DB logic to `internal/model`
- Add unit tests
- Add database migrations
- Add more business logic and endpoints

---

**This project is a work in progress.**
