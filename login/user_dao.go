package login

import (
	"reflect"

	"gorm.io/gorm"
)

type userDao struct {
	db    *gorm.DB
	tUser reflect.Type
}

func (d *userDao) getUserByID(id string) (user interface{}, err error) {
	user = reflect.New(d.tUser).Interface()
	err = d.db.Where("id = ?", id).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userDao) getUserByOAuthUserID(provider string, oid string) (user interface{}, err error) {
	user = reflect.New(d.tUser).Interface()
	err = d.db.Where("o_auth_provider = ? and o_auth_user_id = ?", provider, oid).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userDao) getUserByUsername(name string) (user interface{}, err error) {
	user = reflect.New(d.tUser).Interface()
	err = d.db.Where("username = ?", name).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userDao) getUserByOAuthIndentifier(provider string, indentifier string) (user interface{}, err error) {
	user = reflect.New(d.tUser).Interface()
	err = d.db.Where("o_auth_provider = ? and o_auth_indentifier = ?", provider, indentifier).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userDao) updateOAuthUserID(id string, oid string) (user interface{}, err error) {
	user = reflect.New(d.tUser).Interface()
	err = d.db.Model(user).
		Where("id=?", id).
		Updates(map[string]interface{}{
			"o_auth_user_id": oid,
		}).
		Error
	if err != nil {
		return nil, err
	}
	return d.getUserByID(id)
}
