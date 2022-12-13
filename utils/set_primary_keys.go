package utils

import (
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

func SetPrimaryKeys(from, to interface{}, db *gorm.DB, paramId string) (err error) {
	stmt := &gorm.Statement{DB: db}
	if err = stmt.Parse(to); err != nil {
		return
	}

	for _, v := range stmt.Schema.PrimaryFields {
		if v.Name == "Version" {
			if _, err = to.(interface {
				CreateVersion(db *gorm.DB, paramID string, obj interface{}) (string, error)
			}).CreateVersion(db, paramId, to); err != nil {
				return
			}
			continue
		}
		var value interface{}
		value, err = reflectutils.Get(from, v.Name)
		if err != nil {
			return
		}
		err = reflectutils.Set(to, v.Name, value)
		if err != nil {
			return
		}
	}
	return
}
