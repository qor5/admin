package admin

import (
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get("DB_PARAMS", "admin example database connection string", "user=docs password=docs dbname=docs sslmode=disable host=localhost port=6532 TimeZone=Asia/Tokyo")

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	return db
}
