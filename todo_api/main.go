// Main.go file which initializes mongodb and API
package main

import (
	"todo_api/auth"
	"todo_api/db"
	"todo_api/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize mongodb connection
	db.DbInit()
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/register", auth.Register)
	//curl -X POST http://localhost:8080/register \
	//  -H "Content-Type: application/json" \
	//  -d '{"Username":"jinx","Email":"jinx11@gmail.com","Password":"J1NX"}'
	router.GET("/login", auth.Login)

	// Protected routes(requires auth)
	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware())

	protected.POST("/todos", handler.CreateTodo)
	// curl -X POST http://localhost:8080/todos \
	//-H "Content-Type: application/json" \
	//    -H "Authorization: SjFOWA=="  \
	//    -d '{"username":"jinx","title":"groceries","description":"buy groceries"}'
	protected.GET("/todos", handler.GetTodos)
	// curl -X GET http://localhost:8080/todos \
	//    -H "Content-Type: application/json" \
	//    -H "Authorization: SjFOWA=="  \
	//	  -d '{"username":"jinx"}'
	protected.DELETE("/todos/:id", handler.DeleteTodo)
	// curl -X DELETE http://localhost:8080/todos/1 \
	//    -H "Content-Type: application/json" \
	//    -H "Authorization: SjFOWA=="  \
	//    -d '{"username":"jinx"}'
	protected.PUT("/todos/:id", handler.UpdateTodo)
	//# Update a todo with ID 1
	//curl -X PUT http://localhost:8080/todos/1 \
	//  -H "Content-Type: application/json" \
	//  -H "Authorization: SjFOWA=="  \
	//  -d '{"username":"jinx","title":"Pay bills","description":"Pay electricity bill"}'

	router.Run() // listen and serve on 0.0.0.0:8080
}
