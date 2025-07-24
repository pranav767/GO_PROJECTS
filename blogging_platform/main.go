// Main.go file which initializes mongodb and API
package main

import (
	"blogging_platform/db"
	"blogging_platform/handler"
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

	router.POST("/posts", handler.CreateBlogPost)
	// curl -X POST http://localhost:8080/posts   -H "Content-Type: application/json"   -d '{"title":"My First blog Post","content":"This is the content of my first blog post.","category":"Technology","tags":["Tech","Programming"]}'
	router.GET("/posts", handler.GetBlogPost)
	// curl -X GET http://localhost:8080/posts
	router.DELETE("/posts/:id", handler.DeleteBlogPost)
	//curl -X DELETE http://localhost:8080/posts/2
	router.PUT("/posts/:id", handler.UpdateBlogPost)
	//# Update a blog post with ID 1
	//curl -X PUT http://localhost:8080/posts/1 \
	//  -H "Content-Type: application/json" \
	//  -d '{"title":"Updated Title","content":"Updated content","category":"Updated Category","tags":["updated","new"]}'

	router.Run() // listen and serve on 0.0.0.0:8080
}
