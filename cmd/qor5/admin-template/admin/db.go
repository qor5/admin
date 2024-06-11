package admin

import (
	"github.com/qor5/admin/v3/cmd/qor5/admin-template/models"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get("DB_PARAMS", "database connection string", "")

func ConnectDB() (db *gorm.DB) {
	var err error
	// Create database connection
	db, err = gorm.Open(postgres.Open(dbParamsString))
	if err != nil {
		panic(err)
	}

	// Set db log level
	db.Logger = db.Logger.LogMode(logger.Info)

	// Create data table in the database
	err = db.AutoMigrate(models.Post{})
	if err != nil {
		panic(err)
	}

	return
}
