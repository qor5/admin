package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	Name        string
	Permissions pq.StringArray `gorm:"type:text[]"`
}
