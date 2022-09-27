package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/login"
	"github.com/ua-parser/uap-go/uaparser"
	"gorm.io/gorm"
)

const (
	LoginTokenHashLen = 8 // The hash string length of the token stored in the DB.
)

func addSessionLogByUserID(r *http.Request, userID uint) (err error) {
	c, err := r.Cookie(authCookieName)
	if err != nil {
		return err
	}

	client := uaparser.NewFromSaved().Parse(r.Header.Get("User-Agent"))

	if err = db.Model(&models.LoginSession{}).Create(&models.LoginSession{
		UserID:    userID,
		Device:    fmt.Sprintf("%v - %v", client.UserAgent.Family, client.Os.Family),
		IP:        ip(r),
		TokenHash: getStringHash(c.Value, LoginTokenHashLen),
		ExpiredAt: time.Now().Add(time.Duration(loginBuilder.GetSessionMaxAge()) * time.Second),
	}).Error; err != nil {
		return err
	}

	return nil
}

func updateCurrentSessionLog(r *http.Request, userID uint) (err error) {
	c, err := r.Cookie(authCookieName)
	if err != nil {
		return err
	}
	tokenHash := getStringHash(c.Value, LoginTokenHashLen)
	oldTokenHash := getStringHash(login.GetOldSessionTokenBeforeExtend(r), LoginTokenHashLen)
	if err = db.Model(&models.LoginSession{}).
		Where("user_id = ? and token_hash = ?", userID, oldTokenHash).
		Updates(map[string]interface{}{
			"token_hash": tokenHash,
			"expired_at": time.Now().Add(time.Duration(loginBuilder.GetSessionMaxAge()) * time.Second),
		}).Error; err != nil {
		return err
	}

	return nil
}

func expireCurrentSessionLog(r *http.Request, userID uint) (err error) {
	c, err := r.Cookie(authCookieName)
	if err != nil {
		return err
	}
	tokenHash := getStringHash(c.Value, LoginTokenHashLen)
	if err = db.Model(&models.LoginSession{}).
		Where("user_id = ? and token_hash = ?", userID, tokenHash).
		Updates(map[string]interface{}{
			"expired_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	return nil
}

func expireAllSessionLogs(userID uint) (err error) {
	return db.Model(&models.LoginSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"expired_at": time.Now(),
		}).Error
}

func expireOtherSessionLogs(r *http.Request, userID uint) (err error) {
	sc, _ := r.Cookie(authCookieName)

	return db.Model(&models.LoginSession{}).
		Where("user_id = ? AND token_hash != ?", userID, getStringHash(sc.Value, LoginTokenHashLen)).
		Updates(map[string]interface{}{
			"expired_at": time.Now(),
		}).Error
}

func isTokenValid(v models.LoginSession) bool {
	return time.Now().Sub(v.ExpiredAt) > 0
}

func checkIsTokenValidFromRequest(r *http.Request, userID uint) (valid bool, err error) {
	sc, err := r.Cookie(authCookieName)
	if err != nil {
		return false, nil
	}
	sessionLog := models.LoginSession{}
	if err = db.Where("user_id = ? and token_hash = ?", userID, getStringHash(sc.Value, LoginTokenHashLen)).
		First(&sessionLog).
		Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}
		return false, nil
	}
	// IP check
	if sessionLog.IP != ip(r) {
		return false, nil
	}
	if isTokenValid(sessionLog) {
		return false, nil
	}
	return true, nil
}
