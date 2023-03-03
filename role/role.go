package role

import (
	"github.com/qor5/x/perm"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	Name        string                  `gorm:"unique"`
	Permissions []*perm.DefaultDBPolicy `gorm:"foreignKey:ReferID"`
}
