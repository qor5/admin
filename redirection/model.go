package redirection

import (
	"gorm.io/gorm"
)

type (
	ObjectRedirection struct {
		gorm.Model
		Source string `csv:"source"`
		Target string `csv:"target"`
	}
)

func (b *ObjectRedirection) TableName() string {
	return "object_redirections"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ObjectRedirection{})
}
