package login

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

type SessionSecurer interface {
	UpdateSecure(db *gorm.DB, model interface{}, id string) error
	GetSecure() string
}

type SessionSecure struct {
	SessionSecure string `gorm:"size:32"`
}

var _ SessionSecurer = (*SessionSecure)(nil)

func (ss *SessionSecure) UpdateSecure(db *gorm.DB, model interface{}, id string) error {
	b := make([]byte, 16)
	rand.Read(b)
	secure := hex.EncodeToString(b)
	err := db.Model(model).
		Where(fmt.Sprintf("%s = ?", snakePrimaryField(model)), id).
		Updates(map[string]interface{}{
			"session_secure": secure,
		}).
		Error
	if err != nil {
		return err
	}
	ss.SessionSecure = secure
	return nil
}

func (ss *SessionSecure) GetSecure() string {
	return ss.SessionSecure
}
