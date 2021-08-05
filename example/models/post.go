package models

import (
	"time"

	"github.com/qor/media/media_library"
)

type Post struct {
	ID        uint
	Title     string
	Body      string
	HeroImage media_library.MediaBox `sql:"type:text;"`
	UpdatedAt time.Time
	CreatedAt time.Time
}
