package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/v3/presets/examples"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get("DB_PARAMS", "presets example database connection string", "")

func main() {
	db, err := gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.Logger.LogMode(logger.Info)

	p := examples.Preset1(db)

	log.Println("serving on :7001")
	log.Fatal(http.ListenAndServe(":7001", middleware.Logger(
		middleware.RequestID(p))))
}
