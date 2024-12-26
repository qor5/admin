package main

import (
	"fmt"
	"net/http"

	"github.com/qor5/admin/v3/emailbuilder"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	mux := http.NewServeMux()
	db := ConnectDB()
	eb := emailbuilder.ConfigEmailBuilder(db)
	mux.Handle("/email_template/", http.StripPrefix("/email_template", eb))
	fmt.Println("Listen on http://localhost:9800")
	err := http.ListenAndServe(":9800", mux)
	if err != nil {
		panic(err)
	}
}

var dbParamsString = osenv.Get("DB_PARAMS", "email builder example database connection string", "")

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}
