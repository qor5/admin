package role

import (
	"github.com/goplaid/x/perm"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	Name        string
	Permissions []*perm.DefaultDBPolicy `gorm:"foreignKey:ReferID"`
}
