package login

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type UserPasser interface {
	EncryptPassword()
	IsPasswordCorrect(password string) bool
}

type UserPass struct {
	Username string `gorm:"index:uidx_users_username,unique,where:username!=''"`
	Password string
	Salt     string
}

func (up *UserPass) EncryptPassword() {
	if up.Password == "" {
		return
	}
	up.Salt = up.genSalt()
	up.Password = up.doEncryptPassword(up.Password, up.Salt)
}

func (up *UserPass) doEncryptPassword(password string, salt string) string {
	sum := sha256.Sum256([]byte(salt + password))
	return hex.EncodeToString(sum[:])
}

func (up *UserPass) IsPasswordCorrect(password string) bool {
	ep := up.doEncryptPassword(password, up.Salt)
	return ep == up.Password
}

func (up *UserPass) genSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type OAuthUser interface {
	SetAvatar(v string)
}

type OAuthInfo struct {
	OAuthProvider string `gorm:"index:uidx_users_oauth,unique,where:o_auth_provider!='' and o_auth_user_id!=''"`
	OAuthUserID   string `gorm:"index:uidx_users_oauth,unique,where:o_auth_provider!='' and o_auth_user_id!=''"`
	// the value that user can get to indentify his account
	// in most cases is email or account name
	// it is used to find the user record on the first login
	OAuthIndentifier string
	OAuthAvatar      string `gorm:"-"`
}

func (oa *OAuthInfo) SetAvatar(v string) {
	oa.OAuthAvatar = v
}
