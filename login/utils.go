package login

import (
	"net/http"

	"gorm.io/gorm"
)

func RevokeTOTP(
	ssup SessionSecureUserPasser,
	db *gorm.DB,
	userModel interface{},
	userID string,
) (err error) {
	if err = ssup.SetIsTOTPSetup(db, userModel, false); err != nil {
		return err
	}
	if err = ssup.UpdateSecure(db, userModel, userID); err != nil {
		return err
	}
	return nil
}

func GetSessionToken(b *Builder, r *http.Request) string {
	c, err := r.Cookie(b.authCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
