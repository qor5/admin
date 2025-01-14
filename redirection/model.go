package redirection

import (
	"gorm.io/gorm"
)

type (
	ObjectRedirection struct {
		gorm.Model
		Source string
		Target string
	}
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ObjectRedirection{})
}
