package main

import (
	"fmt"

	"github.com/qor5/admin/example/admin"
	"gorm.io/gorm"
)

func main() {
	db := admin.ConnectDB()
	emptyDB(db)
	initDB(db)

	return
}

func emptyDB(db *gorm.DB) {
	var tableNames []string

	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public';").Scan(&tableNames).
		Error; err != nil {
		panic(err)
	}

	for _, name := range tableNames {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", name)).Error; err != nil {
			panic(err)
		}
	}
}

func initDB(db *gorm.DB) {
	admin.GenInitialPasswordUser()
	// TODO: import init data.
}
