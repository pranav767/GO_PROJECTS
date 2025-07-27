# Todo List API

This project is a simple **Todo List API** built using the Go programming language and MongoDB as the database. It provides endpoints for creating, retrieving, updating, and deleting todos. The project is inspired by the [Todo List API project on roadmap.sh](https://roadmap.sh/projects/todo-list-api).

## Features

- **User Authentication**: Secure authentication using tokens.
- **CRUD Operations**: Create, Read, Update, and Delete todos.
- **Pagination**: Retrieve todos with pagination support.
- **Error Handling**: Standardized error responses.

## Endpoints

### Authentication

#### Register
- **Endpoint**: `POST /register`
- **Description**: Register a new user.
- **Request Body**:
  ```json
  {
    "username": "jinx",
    "email": "jinx11@gmail.com",
    "password": "J1NX"
  }
  ```

#### Login
- **Endpoint**: `GET /login`
- **Description**: Login and retrieve an authentication token.

### Todos

#### Create Todo
- **Endpoint**: `POST /todos`
- **Description**: Create a new todo.
- **Request Body**:
  ```json
  {
    "title": "Buy groceries",
    "description": "Buy milk, eggs, bread"
  }
  ```

#### Get Todos
- **Endpoint**: `GET /todos`
- **Description**: Retrieve all todos for the authenticated user with pagination.
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `limit`: Number of items per page (default: 10)

#### Update Todo
- **Endpoint**: `PUT /todos/:id`
- **Description**: Update an existing todo.
- **Request Body**:
  ```json
  {
    "title": "Pay bills",
    "description": "Pay electricity and water bills"
  }
  ```

#### Delete Todo
- **Endpoint**: `DELETE /todos/:id`
- **Description**: Delete a todo by ID.

## Project Structure

```
.
├── auth
│   ├── auth.go          # Authentication middleware and logic
├── db
│   ├── mongo.go         # MongoDB connection setup
├── handler
│   ├── handler.go       # API endpoint handlers
│   ├── utils.go         # Utility functions
├── models
│   ├── todo_list.go     # Data models for todos and users
├── main.go              # Entry point of the application
```

## How to Run

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd todo-list-api
   ```

2. **Set Up MongoDB**:
   - Ensure MongoDB is running locally or provide a connection string in the `db/mongo.go` file.

3. **Run the Application**:
   ```bash
   go run main.go
   ```

4. **Test the Endpoints**:
   Use tools like `curl` or Postman to test the API endpoints.

## Additional Setup for MongoDB

To run MongoDB as a container, you can use Docker. Follow these steps:

### Run MongoDB with Docker

1. **Pull the MongoDB Docker Image**:
   ```bash
   docker pull mongo:latest
   ```

2. **Run the MongoDB Container**:
   ```bash
   docker run -d \
     --name mongodb \
     -p 27017:27017 \
     -e MONGO_INITDB_ROOT_USERNAME=admin \
     -e MONGO_INITDB_ROOT_PASSWORD=adminpass \
     mongo:latest
   ```

3. **Verify the Container is Running**:
   ```bash
   docker ps
   ```

### Interact with MongoDB

You can use the `mongosh` shell to interact with the MongoDB instance:

```bash
docker exec -it mongodb mongosh -u admin -p adminpass
```

## MongoDB Commands

Here are some useful MongoDB commands to interact with the database:

### Use the Database
```bash
use todo_db
```

### Show Collections
```bash
show collections
```

### Check if Documents Exist
To check if documents exist in a collection:
```bash
db.todos.find()
```

### Delete a Document
To delete a specific document by its ID:
```bash
db.todos.deleteOne({ _id: ObjectId("<document_id>") })
```

### Count Documents
To count the number of documents in a collection:
```bash
db.todos.countDocuments()
```

### Drop a Collection
To delete an entire collection:
```bash
db.todos.drop()
```

## Example Requests

### Create a Todo
```bash
curl -X POST http://localhost:8080/todos \
     -H "Content-Type: application/json" \
     -H "Authorization: SjFOWA==" \
     -d '{"title":"Buy groceries","description":"Buy milk, eggs, bread"}'
```

### Get Todos with Pagination
```bash
curl -X GET "http://localhost:8080/todos?page=1&limit=10" \
     -H "Content-Type: application/json" \
     -H "Authorization: SjFOWA=="
```

## Additional Example Requests

### Register a User
```bash
curl -X POST http://localhost:8080/register \
     -H "Content-Type: application/json" \
     -d '{"username":"jinx","email":"jinx11@gmail.com","password":"J1NX"}'
```

### Login
```bash
curl -X GET http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"username":"jinx","password":"J1NX"}'
```

### Update a Todo
```bash
curl -X PUT http://localhost:8080/todos/1 \
     -H "Content-Type: application/json" \
     -H "Authorization: SjFOWA==" \
     -d '{"title":"Pay bills","description":"Pay electricity and water bills"}'
```

### Delete a Todo
```bash
curl -X DELETE http://localhost:8080/todos/1 \
     -H "Content-Type: application/json" \
     -H "Authorization: SjFOWA=="
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Acknowledgments

- Inspired by the [Todo List API project on roadmap.sh](https://roadmap.sh/projects/todo-list-api).
