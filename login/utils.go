package login

import (
	"net/http"

	"gorm.io/gorm"
)

// renewSession will reset cookie with the latest user info in DB.
func renewSession(
	b *Builder,
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	claims, err := parseUserClaimsFromCookie(r, b.authCookieName, b.secret)
	if err != nil {
		return
	}

	var user interface{}
	user, err = b.findUserByID(claims.UserID)
	if err != nil {
		return
	}

	if err = b.setSecureCookiesByClaims(w, user, *claims); err != nil {
		return
	}
	return nil
}

func SignOutAllOtherSessions(
	b *Builder,
	w http.ResponseWriter,
	r *http.Request,
	db *gorm.DB,
	userModel interface{},
	userID string,
	ss SessionSecurer,
) (err error) {
	if err = ss.UpdateSecure(db, userModel, userID); err != nil {
		return err
	}
	if err = renewSession(b, w, r); err != nil {
		return err
	}
	return nil
}

func RevokeTOTP(
	db *gorm.DB,
	userModel interface{},
	userID string,
	ss SessionSecurer,
	up UserPasser,
) (err error) {
	if err = up.SetIsTOTPSetup(db, userModel, false); err != nil {
		return err
	}
	if err = ss.UpdateSecure(db, userModel, userID); err != nil {
		return err
	}
	return nil
}
