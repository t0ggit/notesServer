package dto

type Note struct {
	ID       int64  `json:"id"`
	Name     string `json:"name,omitempty"`
	LastName string `json:"last_name,omitempty"`
	Content  string `json:"note,omitempty"`
}

func NewNote() *Note {
	return &Note{ID: -1}
}
