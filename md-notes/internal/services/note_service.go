package services

import (
	"md-notes/models"
	"os"
)

func SaveNoteToFile(note models.Note) error {
	file, err := os.Create("notes/" + note.ID + ".md")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(note.Content)
	return err
}
