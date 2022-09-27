package login

import (
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
