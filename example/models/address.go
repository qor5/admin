package models

import (
	"time"

	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

type Address struct {
	gorm.Model

	Street    string
	HomeImage media_library.MediaBox `sql:"type:text;"`
	UpdatedAt time.Time
	CreatedAt time.Time

	publish.Status
	Phones []*Phone
}

type Phone struct {
	gorm.Model
	AddressID uint
	Number    int
}
