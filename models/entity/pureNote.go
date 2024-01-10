package entity

import "notesServer/models/dto"

// PureNote это тот же dto.Note, но без ID. Это нужно для хранения заметок в storage.
type PureNote struct {
	name     string
	lastName string
	content  string
}

// GetPureNote преобразует dto.Note в PureNote
func GetPureNote(note *dto.Note) PureNote {
	return PureNote{
		name:     note.Name,
		lastName: note.LastName,
		content:  note.Content,
	}
}

// ToNoteWithID возвращает dto.Note с указанным ID
func (pn PureNote) ToNoteWithID(id int64) *dto.Note {
	return &dto.Note{
		ID:       id,
		Name:     pn.name,
		LastName: pn.lastName,
		Content:  pn.content,
	}
}
