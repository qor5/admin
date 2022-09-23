package note

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

func (this *QorNote) BeforeCreate(tx *gorm.DB) (err error) {
	if strings.TrimSpace(this.Content) == "" {
		err = errors.New("Note cannot be empty")
	}

	return
}

type UserNote struct {
	gorm.Model

	UserID       uint   `gorm:"index"`
	ResourceType string `gorm:"index"`
	ResourceID   string `gorm:"index"`
	Number       int64
}

func GetUnreadNotesCount(db *gorm.DB, userID uint, resourceType, resourceID string) int64 {
	var total int64
	db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Count(&total)

	if total == 0 {
		return 0
	}

	userNote := UserNote{}
	db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", userID, resourceType, resourceID).First(&userNote)
	return total - userNote.Number
}
