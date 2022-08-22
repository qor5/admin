package login

import "gorm.io/gorm"

type OAuthUser interface {
	FindUserByOAuthUserID(db *gorm.DB, model interface{}, provider string, oid string) (user interface{}, err error)
	FindUserByOAuthIndentifier(db *gorm.DB, model interface{}, provider string, indentifier string) (user interface{}, err error)
	InitOAuthUserID(db *gorm.DB, model interface{}, provider string, indentifier string, oid string) error
	SetAvatar(v string)
	GetAvatar() string
}

type OAuthInfo struct {
	OAuthProvider string `gorm:"index:uidx_users_oauth_provider_user_id,unique,where:o_auth_provider!='' and o_auth_user_id!='';index:uidx_users_oauth_provider_indentifier,unique,where:o_auth_provider!='' and o_auth_indentifier!=''"`
	OAuthUserID   string `gorm:"index:uidx_users_oauth_provider_user_id,unique,where:o_auth_provider!='' and o_auth_user_id!=''"`
	// users use this value to log into their account
	// in most cases is email or account name
	// it is used to find the user record on the first login
	OAuthIndentifier string `gorm:"index:uidx_users_oauth_provider_indentifier,unique,where:o_auth_provider!='' and o_auth_indentifier!=''"`
	OAuthAvatar      string `gorm:"-"`
}

var _ OAuthUser = (*OAuthInfo)(nil)

func (oa *OAuthInfo) FindUserByOAuthUserID(db *gorm.DB, model interface{}, provider string, oid string) (user interface{}, err error) {
	err = db.Where("o_auth_provider = ? and o_auth_user_id = ?", provider, oid).
		First(model).
		Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (oa *OAuthInfo) FindUserByOAuthIndentifier(db *gorm.DB, model interface{}, provider string, indentifier string) (user interface{}, err error) {
	err = db.Where("o_auth_provider = ? and o_auth_indentifier = ?", provider, indentifier).
		First(model).
		Error
	if err != nil {
		return nil, err
	}
	return model, nil
}
func (oa *OAuthInfo) InitOAuthUserID(db *gorm.DB, model interface{}, provider string, indentifier string, oid string) error {
	err := db.Model(model).
		Where("o_auth_provider = ? and o_auth_indentifier = ?", provider, indentifier).
		Updates(map[string]interface{}{
			"o_auth_user_id": oid,
		}).
		Error
	if err != nil {
		return err
	}
	oa.OAuthUserID = oid
	return nil
}

func (oa *OAuthInfo) SetAvatar(v string) {
	oa.OAuthAvatar = v
}

func (oa *OAuthInfo) GetAvatar() string {
	return oa.OAuthAvatar
}
