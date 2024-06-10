package activity

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type QorNote struct {
	gorm.Model
	UserID       uint `gorm:"index"`
	Creator      string
	ResourceType string `gorm:"index"`
	ResourceID   string `gorm:"index"`
	Content      string `sql:"size:5000"`
}

func (n *QorNote) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(n.Content) == "" {
		return errors.New("note cannot be empty")
	}
	return nil
}

type UserNote struct {
	gorm.Model
	UserID       uint   `gorm:"index"`
	ResourceType string `gorm:"index"`
	ResourceID   string `gorm:"index"`
	Number       int64
}
