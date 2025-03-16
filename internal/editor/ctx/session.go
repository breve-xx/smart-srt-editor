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
