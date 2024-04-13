package role

import (
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	Name        string                  `gorm:"unique"`
	Permissions []*perm.DefaultDBPolicy `gorm:"foreignKey:ReferID"`
}
