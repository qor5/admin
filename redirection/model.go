package redirection

import (
	"gorm.io/gorm"
)

type (
	Redirection struct {
		gorm.Model
		Source string `csv:"source"`
		Target string `csv:"target"`
	}
)

func (*Redirection) TableName() string {
	return "redirections"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Redirection{})
}
