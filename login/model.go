package login

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserPasser interface {
	EncryptPassword()
	IsPasswordCorrect(password string) bool
	GetPassUpdatedAt() string
}

type UserPass struct {
	Username string `gorm:"index:uidx_users_username,unique,where:username!=''"`
	Password string `gorm:"size:60"`
	// UnixNano string
	PassUpdatedAt string
}

func (up *UserPass) EncryptPassword() {
	if up.Password == "" {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(up.Password), 10)
	if err != nil {
		panic(err)
	}
	up.Password = string(hash)
	up.PassUpdatedAt = fmt.Sprint(time.Now().UnixNano())
}

func (up *UserPass) IsPasswordCorrect(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(up.Password), []byte(password)) == nil
}

func (up *UserPass) GetPassUpdatedAt() string {
	return up.PassUpdatedAt
}

type OAuthUser interface {
	SetAvatar(v string)
}

type OAuthInfo struct {
	OAuthProvider string `gorm:"index:uidx_users_oauth_provider_user_id,unique,where:o_auth_provider!='' and o_auth_user_id!='';index:uidx_users_oauth_provider_indentifier,unique,where:o_auth_provider!='' and o_auth_indentifier!=''"`
	OAuthUserID   string `gorm:"index:uidx_users_oauth_provider_user_id,unique,where:o_auth_provider!='' and o_auth_user_id!=''"`
	// the value that user can get to indentify his account
	// in most cases is email or account name
	// it is used to find the user record on the first login
	OAuthIndentifier string `gorm:"index:uidx_users_oauth_provider_indentifier,unique,where:o_auth_provider!='' and o_auth_indentifier!=''"`
	OAuthAvatar      string `gorm:"-"`
}

func (oa *OAuthInfo) SetAvatar(v string) {
	oa.OAuthAvatar = v
}
