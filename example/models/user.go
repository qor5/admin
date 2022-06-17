package models

import (
	"github.com/qor/qor5/role"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name      string
	Company   string
	Email     string
	Roles     []role.Role `gorm:"many2many:user_role_join;"`
	Status    string
	AvatarURL string
	UpdatedAt time.Time
	CreatedAt time.Time

	OAuthProvider string `gorm:"index:uidx_users_oauth,unique"`
	OAuthUserID   string `gorm:"index:uidx_users_oauth,unique"`
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
