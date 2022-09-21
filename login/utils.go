package login

import (
	"net/http"

	"gorm.io/gorm"
)

func SignOutAllOtherSessions(
	b *Builder,
	w http.ResponseWriter,
	r *http.Request,
	ss SessionSecurer,
	db *gorm.DB,
	userModel interface{},
	userID string,
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
