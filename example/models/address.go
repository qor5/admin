package models

import (
	"time"

	"github.com/lib/pq"
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
	Phones pq.StringArray `gorm:"type:varchar(100)[]"`
}
