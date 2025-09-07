package models

type Note struct {
	ID       string `json:"id"`
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content"`
}
