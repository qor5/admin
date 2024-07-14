package activity

import (
	"time"

	"gorm.io/gorm"
)

type ActivityUser struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"index;not null"`
	Avatar    string         `gorm:"not null"`
}
