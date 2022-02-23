package models

import (
	"github.com/qor/qor5/media/media_library"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model

	Code  string
	Name  string
	Image media_library.MediaBox `sql:"type:text;"`
}
