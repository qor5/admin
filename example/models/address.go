package models

import (
	"time"

	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/publish"
)

type Customer struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Addresses []*Address
}

type Address struct {
	ID         uint `gorm:"primarykey"`
	CustomerID uint

	Street    string
	HomeImage media_library.MediaBox `sql:"type:text;"`
	UpdatedAt time.Time
	CreatedAt time.Time

	publish.Status
	Phones []*Phone
}

type Phone struct {
	ID        uint `gorm:"primarykey"`
	AddressID uint
	Number    int
}
