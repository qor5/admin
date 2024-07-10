package activity

import "gorm.io/gorm"

type ActivityUser struct {
	gorm.Model
	Name   string `gorm:"index"`
	Avatar string
}
