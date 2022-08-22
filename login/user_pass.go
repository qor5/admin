package login

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserPasser interface {
	FindUser(db *gorm.DB, model interface{}, username string) (user interface{}, err error)
	GetUsername() string
	EncryptPassword()
	IsPasswordCorrect(password string) bool
	GetPasswordUpdatedAt() string
}

type UserPass struct {
	Username string `gorm:"index:uidx_users_username,unique,where:username!=''"`
	Password string `gorm:"size:60"`
	// UnixNano string
	PassUpdatedAt string
}

var _ UserPasser = (*UserPass)(nil)

func (up *UserPass) FindUser(db *gorm.DB, model interface{}, username string) (user interface{}, err error) {
	err = db.Where("username = ?", username).
		First(model).
		Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (up *UserPass) GetUsername() string {
	return up.Username
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

func (up *UserPass) GetPasswordUpdatedAt() string {
	return up.PassUpdatedAt
}
