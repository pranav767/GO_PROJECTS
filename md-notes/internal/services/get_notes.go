package services

import (
	"md-notes/models"
	"os"
)

func GetAllNotes() ([]models.Note, error) {
	// create a variable to hold structure of notes
	var notes []models.Note
	files, err := os.ReadDir("notes")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		note, err := LoadNoteFromFile(file.Name())
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func LoadNoteFromFile(filename string) (models.Note, error) {
	var note models.Note
	data, err := os.ReadFile("notes/" + filename)
	if err != nil {
		return note, err
	}
	note.Filename = filename
	note.Content = string(data)
	// Optionally set ID from filename (strip .md)
	if len(filename) > 3 && filename[len(filename)-3:] == ".md" {
		note.ID = filename[:len(filename)-3]
	} else {
		note.ID = filename
	}
	// CreatedAt can be set using file info if needed
	return note, nil
}
