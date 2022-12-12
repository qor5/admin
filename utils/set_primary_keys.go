package utils

import (
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

func SetPrimaryKeys(from, to interface{}, db *gorm.DB, paramId string) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(to); err != nil {
		panic(err)
	}

	for _, v := range stmt.Schema.PrimaryFields {
		if v.Name == "Version" {
			to.(interface {
				CreateVersion(db *gorm.DB, paramID string, obj interface{}) string
			}).CreateVersion(db, paramId, to)
			continue
		}
		value, err := reflectutils.Get(from, v.Name)
		if err != nil {
			panic(err)
		}
		err = reflectutils.Set(to, v.Name, value)
		if err != nil {
			panic(err)
		}
	}
}
