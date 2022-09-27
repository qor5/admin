package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/login"
	"github.com/ua-parser/uap-go/uaparser"
)

const (
	LoginTokenHashLen = 8 // The hash string length of the token stored in the DB.
)

func addSessionLogByUserID(r *http.Request, userID uint) (err error) {
	c, err := r.Cookie(AuthCookieName)
	if err != nil {
		return err
	}

	client := uaparser.NewFromSaved().Parse(r.Header.Get("User-Agent"))

	if err = db.Model(&models.LoginSession{}).Create(&models.LoginSession{
		UserID:    userID,
		Device:    fmt.Sprintf("%v - %v", client.UserAgent.Family, client.Os.Family),
		IP:        r.RemoteAddr,
		TokenHash: getStringHash(c.Value, LoginTokenHashLen),
	}).Error; err != nil {
		return err
	}

	return nil
}

func delCurrentSessionLog(r *http.Request) (err error) {
	c, err := r.Cookie(AuthCookieName)
	if err != nil {
		return err
	}

	tokenHash := getStringHash(c.Value, LoginTokenHashLen)
	if err = db.Delete(&models.LoginSession{}, "token_hash = ?", tokenHash).Error; err != nil {
		return err
	}

	return nil
}

func isTokenExpired(b *login.Builder, v models.LoginSession) bool {
	return time.Now().Sub(v.CreatedAt) > time.Duration(b.GetSessionMaxAge())*time.Second
}

func isUnexpiredTokenInvalid(r *http.Request, b *login.Builder) (err error, ok bool) {
	c, err := r.Cookie(AuthCookieName)
	if err != nil {
		return err, false
	}

	sessionItems := []*models.LoginSession{}
	// If this token has been deleted , indicates that this token is invalid,
	// such as this session token has been signed out by the user actively.
	if err = db.Model(&models.LoginSession{}).Unscoped().
		Where("token_hash = ? AND deleted_at IS NOT NULL", getStringHash(c.Value, LoginTokenHashLen)).
		Find(&sessionItems).Error; err != nil {
		return err, false
	}

	// Exclude expired token.
	var cnt int64
	for _, s := range sessionItems {
		if !isTokenExpired(b, *s) {
			cnt++
		}
	}

	// cnt represents the number of token that has not expired but has been deleted.
	return nil, cnt > 0
}

// emptyOtherSessionLog will delete other session logs by userID.
func emptyOtherSessionLog(r *http.Request, userID uint) (err error) {
	c, err := r.Cookie(AuthCookieName)
	if err != nil {
		return err
	}

	if err = db.Delete(
		&models.LoginSession{},
		"user_id = ? AND token_hash != ?", userID, getStringHash(c.Value, LoginTokenHashLen)).
		Error; err != nil {
		return err
	}

	return nil
}
