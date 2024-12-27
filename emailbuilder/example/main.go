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

	withCORS := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	db := ConnectDB()
	eb := emailbuilder.ConfigEmailBuilder(db)
	mux.Handle("/email_template/", http.StripPrefix("/email_template", eb))

	fmt.Println("Listen on http://localhost:9800")
	if err := http.ListenAndServe(":9800", withCORS(mux)); err != nil {
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
