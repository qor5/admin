package models

import (
	"time"
)

type User struct {
	ID         uint
	Name       string
	Company    string
	Email      string
	Permission string
	Status     string
	AvatarURL  string
	UpdatedAt  time.Time
	CreatedAt  time.Time

	OAuthProvider string `gorm:"index:uidx_users_oauth,unique"`
	OAuthUserID   string `gorm:"index:uidx_users_oauth,unique"`
}
