package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/qor5/admin/presets/examples"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(postgres.Open(os.Getenv("DBString")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.Logger.LogMode(logger.Info)

	p := examples.Preset1(db)

	log.Println("serving on :7001")
	log.Fatal(http.ListenAndServe(":7001", middleware.Logger(
		middleware.RequestID(p))))
}
