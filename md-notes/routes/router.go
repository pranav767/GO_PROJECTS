package routes

import (
	"md-notes/internal/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/upload", controllers.UploadMarkdownFile)
	//router.GET("/note/:id/grammar", controllers.CheckGrammar)
	router.POST("/notes", controllers.UploadMarkdown)
	router.GET("/notes", controllers.GetNotes)
	router.GET("/notes/:id/html", controllers.RenderNoteHTML)
	router.GET("/notes/:id/grammar", controllers.CheckGrammar)

	return router
}
