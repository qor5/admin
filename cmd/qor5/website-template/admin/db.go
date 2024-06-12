package admin

import (
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get(
	"DB_PARAMS",
	"database connection string",
	"user=website password=123 dbname=website_dev sslmode=disable host=localhost port=6432",
)

func ConnectDB() (db *gorm.DB) {
	var err error

	db, err = gorm.Open(postgres.Open(dbParamsString))
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	return
}

type MyContent struct {
	ID    uint
	Text  string
	Color string
}

type MenuItem struct {
	Text string
	Link string
}

type MyHeader struct {
	gorm.Model
	MenuItems []*MenuItem `gorm:"serializer:json;type:json"`
}
