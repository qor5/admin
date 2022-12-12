package utils

import (
	"gorm.io/gorm"
)

func GetPrimaryKeys(obj interface{}, db *gorm.DB) (result []string) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(obj); err != nil {
		panic(err)
	}

	for _, v := range stmt.Schema.PrimaryFields {
		result = append(result, v.DBName)
	}
	return
}
