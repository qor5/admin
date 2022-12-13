package utils

import (
	"gorm.io/gorm"
)

func GetPrimaryKeys(obj interface{}, db *gorm.DB) (result []string, err error) {
	stmt := &gorm.Statement{DB: db}
	if err = stmt.Parse(obj); err != nil {
		return
	}

	for _, v := range stmt.Schema.PrimaryFields {
		result = append(result, v.DBName)
	}
	return
}
