package controllers

import (
	"md-notes/internal/services"
	"md-notes/models"
	"md-notes/utils"

	"github.com/gin-gonic/gin"
)

// RenderNoteHTML returns the HTML version of a Markdown note by ID
func RenderNoteHTML(c *gin.Context) {
	id := c.Param("id")
	filename := id + ".md"
	note, err := services.LoadNoteFromFile(filename)
	if err != nil {
		c.JSON(404, gin.H{"error": "Note not found"})
		return
	}
	html := services.ConvertMarkdownToHTML(note.Content)
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

func UploadMarkdownFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file is received"})
		return
	}
	uniqueID := utils.GenerateID()
	savePath := "notes/" + uniqueID + ".md"
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}
	c.JSON(200, gin.H{"message": "File uploaded successfully", "id": uniqueID, "filename": file.Filename})
}

func UploadMarkdown(c *gin.Context) {
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	note.ID = utils.GenerateID()

	// Save the note to a file
	if err := services.SaveNoteToFile(note); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save note"})
		return
	}

	c.JSON(200, gin.H{"message": "Note uploaded successfully", "id": note.ID})
}

func GetNotes(c *gin.Context) {
	notes, err := services.GetAllNotes()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve notes"})
		return
	}
	c.JSON(200, notes)
}

func CheckGrammar(c *gin.Context) {
	id := c.Param("id")
	filename := id + ".md"
	note, err := services.LoadNoteFromFile(filename)
	if err != nil {
		c.JSON(404, gin.H{"error": "Note not found"})
		return
	}
	grammarErrors := services.CheckGrammar(note.Content)
	c.JSON(200, gin.H{"grammar_errors": grammarErrors})
}
