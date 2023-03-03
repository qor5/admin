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

	// Keep the data of the user table.
	for i, n := range tableNames {
		if n == "users" {
			tableNames = append(tableNames[:i], tableNames[i+1:]...)
			break
		}
	}

	for _, name := range tableNames {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", name)).Error; err != nil {
			panic(err)
		}
	}
}

func initDB(db *gorm.DB) {
	var err error
	// Roles
	admin.InitDefaultRolesToDB(db)
	// Users
	admin.GenInitialPasswordUser()
	// Page Builder
	if err = db.Exec(initPageBuilderSQL).Error; err != nil {
		panic(err)
	}
	// Orders
	if err = db.Exec(initOrdersSQL).Error; err != nil {
		panic(err)
	}
	// Workers
	if err = db.Exec(initWorkersSQL).Error; err != nil {
		panic(err)
	}
	// Categories
	if err = db.Exec(initCategoriesSQL).Error; err != nil {
		panic(err)
	}
	// InputDemos
	if err = db.Exec(initInputDemosSQL).Error; err != nil {
		panic(err)
	}
	// Posts
	if err = db.Exec(initPostsSQL).Error; err != nil {
		panic(err)
	}
	// NestedFieldDemos
	if err = db.Exec(initNestedFieldDemosSQL).Error; err != nil {
		panic(err)
	}
	// ListModels
	if err = db.Exec(initListModelsSQL).Error; err != nil {
		panic(err)
	}
	// MicrositeModels
	if err = db.Exec(initMicrositeModelsSQL).Error; err != nil {
		panic(err)
	}
	// Products
	if err = db.Exec(initProductsSQL).Error; err != nil {
		panic(err)
	}
	// Media Library
	if err = db.Exec(initMediaLibrarySQL).Error; err != nil {
		panic(err)
	}
	// Seq
	if err = db.Exec(initSeqSQL).Error; err != nil {
		panic(err)
	}
}
