package ctx

import (
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/martinlindhe/subtitles"
)

type Session struct {
	ID   uuid.UUID
	File *multipart.FileHeader
	Subs *subtitles.Subtitle
}

func NewSession(ID uuid.UUID, file *multipart.FileHeader, subs *subtitles.Subtitle) *Session {
	return &Session{
		ID:   ID,
		File: file,
		Subs: subs,
	}
}
