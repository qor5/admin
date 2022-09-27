package models

import (
	"gorm.io/gorm"
)

type LoginSession struct {
	gorm.Model

	UserID    uint
	Device    string
	IP        string
	TokenHash string

	Time   string `gorm:"-"`
	Status string `gorm:"-"`
}
