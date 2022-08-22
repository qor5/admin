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
	GetLoginRetryCount() int
	GetLocked() bool
	IncreaseRetryCount(db *gorm.DB, model interface{}, id string) error
	LockUser(db *gorm.DB, model interface{}, id string) error
	UnlockUser(db *gorm.DB, model interface{}, id string) error
}

type UserPass struct {
	Username string `gorm:"index:uidx_users_username,unique,where:username!=''"`
	Password string `gorm:"size:60"`
	// UnixNano string
	PassUpdatedAt   string
	LoginRetryCount int
	Locked          bool
	LockedAt        *time.Time
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

func (up *UserPass) GetLoginRetryCount() int {
	return up.LoginRetryCount
}

func (up *UserPass) GetLocked() bool {
	if !up.Locked {
		return false
	}
	return up.Locked && up.LockedAt != nil && time.Now().Sub(*up.LockedAt) <= time.Hour
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

func (up *UserPass) LockUser(db *gorm.DB, model interface{}, id string) error {
	lockedAt := time.Now()
	if err := db.Model(model).Where(fmt.Sprintf("%s = ?", snakePrimaryField(model)), id).Updates(map[string]interface{}{
		"locked":    true,
		"locked_at": &lockedAt,
	}).Error; err != nil {
		return err
	}

	up.Locked = true
	up.LockedAt = &lockedAt

	return nil
}

func (up *UserPass) UnlockUser(db *gorm.DB, model interface{}, id string) error {
	if err := db.Model(model).Where(fmt.Sprintf("%s = ?", snakePrimaryField(model)), id).Updates(map[string]interface{}{
		"locked":            false,
		"login_retry_count": 0,
		"locked_at":         nil,
	}).Error; err != nil {
		return err
	}

	up.Locked = false
	up.LoginRetryCount = 0
	up.LockedAt = nil

	return nil
}

func (up *UserPass) IncreaseRetryCount(db *gorm.DB, model interface{}, id string) error {
	if err := db.Model(model).Where(fmt.Sprintf("%s = ?", snakePrimaryField(model)), id).Updates(map[string]interface{}{
		"login_retry_count": gorm.Expr("login_retry_count + 1"),
	}).Error; err != nil {
		return err
	}
	up.LoginRetryCount++

	return nil
}
