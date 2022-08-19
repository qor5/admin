package models

import (
	"time"

	"github.com/qor/qor5/login"
	"github.com/qor/qor5/role"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name      string
	Company   string
	Roles     []role.Role `gorm:"many2many:user_role_join;"`
	Status    string
	UpdatedAt time.Time
	CreatedAt time.Time

	// Username is email
	login.UserPass
	login.OAuthInfo
}

func (u User) GetName() string {
	return u.Name
}

func (u User) GetID() uint {
	return u.ID
}

func (u User) GetRoles() (rs []string) {
	for _, r := range u.Roles {
		rs = append(rs, r.Name)
	}
	if len(rs) == 0 {
		rs = []string{"root"}
	}
	return
}
