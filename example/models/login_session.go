package models

import (
	"time"

	"gorm.io/gorm"
)

type LoginSession struct {
	gorm.Model

	UserID    uint `sql:"index"`
	Device    string
	IP        string
	TokenHash string `sql:"index"`
	ExpiredAt time.Time

	Time   string `gorm:"-"`
	Status string `gorm:"-"`
}
