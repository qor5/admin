package login

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserPasser interface {
	FindUser(db *gorm.DB, model interface{}, account string) (user interface{}, err error)
	EncryptPassword()
	IsPasswordCorrect(password string) bool
	IncreaseRetryCount(db *gorm.DB, model interface{}) error
	GenerateResetPasswordToken(db *gorm.DB, model interface{}) (token string, err error)
	ConsumeResetPasswordToken(db *gorm.DB, model interface{}) error

	GetAccountName() string
	GetPasswordUpdatedAt() string
	GetLoginRetryCount() int
	GetLocked() bool
	GetIsTOTPSetup() bool
	GetTOTPSecret() string
	GetLastUsedTOTPCode() (code string, usedAt *time.Time)
	GetResetPasswordToken() (token string, createdAt *time.Time, expired bool)

	SetPassword(db *gorm.DB, model interface{}, password string) error
	SetIsTOTPSetup(db *gorm.DB, model interface{}, v bool) error
	SetTOTPSecret(db *gorm.DB, model interface{}, key string) error
	SetLastUsedTOTPCode(db *gorm.DB, model interface{}, passcode string) error

	LockUser(db *gorm.DB, model interface{}) error
	UnlockUser(db *gorm.DB, model interface{}) error
}

type SessionSecureUserPasser interface {
	SessionSecurer
	UserPasser
}

type UserPass struct {
	Account  string `gorm:"index:uidx_users_account,unique,where:account!='' and deleted_at is null"`
	Password string `gorm:"size:60"`
	// UnixNano string
	PassUpdatedAt               string
	LoginRetryCount             int
	Locked                      bool
	LockedAt                    *time.Time
	ResetPasswordToken          string `gorm:"index:uidx_users_reset_password_token,unique,where:reset_password_token!=''"`
	ResetPasswordTokenCreatedAt *time.Time
	ResetPasswordTokenExpiredAt *time.Time
	TOTPSecret                  string
	IsTOTPSetup                 bool
	LastUsedTOTPCode            string
	LastTOTPCodeUsedAt          *time.Time
}

var _ UserPasser = (*UserPass)(nil)

func (up *UserPass) FindUser(db *gorm.DB, model interface{}, account string) (user interface{}, err error) {
	err = db.Where("account = ?", account).
		First(model).
		Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (up *UserPass) GetAccountName() string {
	return up.Account
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

func (up *UserPass) GetTOTPSecret() string {
	return up.TOTPSecret
}

func (up *UserPass) GetIsTOTPSetup() bool {
	return up.IsTOTPSetup
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

func (up *UserPass) LockUser(db *gorm.DB, model interface{}) error {
	lockedAt := time.Now()
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
		"locked":    true,
		"locked_at": &lockedAt,
	}).Error; err != nil {
		return err
	}

	up.Locked = true
	up.LockedAt = &lockedAt

	return nil
}

func (up *UserPass) UnlockUser(db *gorm.DB, model interface{}) error {
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
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

func (up *UserPass) IncreaseRetryCount(db *gorm.DB, model interface{}) error {
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
		"login_retry_count": gorm.Expr("coalesce(login_retry_count,0) + 1"),
	}).Error; err != nil {
		return err
	}
	up.LoginRetryCount++

	return nil
}

func (up *UserPass) GenerateResetPasswordToken(db *gorm.DB, model interface{}) (token string, err error) {
	token = base64.URLEncoding.EncodeToString([]byte(uuid.NewString()))
	now := time.Now()
	expiredAt := now.Add(10 * time.Minute)
	err = db.Model(model).
		Where("account = ?", up.Account).
		Updates(map[string]interface{}{
			"reset_password_token":            token,
			"reset_password_token_created_at": now,
			"reset_password_token_expired_at": expiredAt,
		}).
		Error
	if err != nil {
		return "", err
	}
	up.ResetPasswordToken = token
	up.ResetPasswordTokenCreatedAt = &now
	up.ResetPasswordTokenExpiredAt = &expiredAt
	return token, nil
}

func (up *UserPass) ConsumeResetPasswordToken(db *gorm.DB, model interface{}) error {
	err := db.Model(model).
		Where("account = ?", up.Account).
		Updates(map[string]interface{}{
			"reset_password_token_expired_at": time.Now(),
		}).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (up *UserPass) GetResetPasswordToken() (token string, createdAt *time.Time, expired bool) {
	if up.ResetPasswordTokenExpiredAt != nil && time.Now().Sub(*up.ResetPasswordTokenExpiredAt) > 0 {
		return "", nil, true
	}
	return up.ResetPasswordToken, up.ResetPasswordTokenCreatedAt, false
}

func (up *UserPass) SetPassword(db *gorm.DB, model interface{}, password string) error {
	up.Password = password
	up.EncryptPassword()
	err := db.Model(model).
		Where("account = ?", up.Account).
		Updates(map[string]interface{}{
			"password":        up.Password,
			"pass_updated_at": up.PassUpdatedAt,
		}).
		Error
	if err != nil {
		return err
	}
	return nil
}

func (up *UserPass) SetTOTPSecret(db *gorm.DB, model interface{}, key string) error {
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
		"totp_secret": key,
	}).Error; err != nil {
		return err
	}

	up.TOTPSecret = key

	return nil
}

func (up *UserPass) SetIsTOTPSetup(db *gorm.DB, model interface{}, v bool) error {
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
		"is_totp_setup": v,
	}).Error; err != nil {
		return err
	}

	up.IsTOTPSetup = v

	return nil
}

func (up *UserPass) SetLastUsedTOTPCode(db *gorm.DB, model interface{}, passcode string) error {
	now := time.Now()
	if err := db.Model(model).Where("account = ?", up.Account).Updates(map[string]interface{}{
		"last_used_totp_code":    passcode,
		"last_totp_code_used_at": &now,
	}).Error; err != nil {
		return err
	}

	up.LastUsedTOTPCode = passcode

	return nil
}

func (up *UserPass) GetLastUsedTOTPCode() (code string, usedAt *time.Time) {
	return up.LastUsedTOTPCode, up.LastTOTPCodeUsedAt
}
