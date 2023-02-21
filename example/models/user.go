package models

import (
	"time"

	"github.com/qor5/admin/role"
	"github.com/qor5/x/login"
	"gorm.io/gorm"
)

const (
	OAuthProviderGoogle          = "google"
	OAuthProviderMicrosoftOnline = "microsoftonline"
	OAuthProviderGithub          = "github"
)

type User struct {
	gorm.Model

	Name             string
	Company          string
	Roles            []role.Role `gorm:"many2many:user_role_join;"`
	Status           string
	UpdatedAt        time.Time
	CreatedAt        time.Time
	FavorPostID      uint
	RegistrationDate time.Time `gorm:"type:date"`

	// Username is email
	login.UserPass
	login.OAuthInfo
	login.SessionSecure
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
