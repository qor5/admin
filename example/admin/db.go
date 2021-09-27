package admin

import (
	"os"

	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/media/media_library"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	if err = db.AutoMigrate(
		&models.Post{},
		&models.InputHarness{},
		&models.User{},
		&media_library.MediaLibrary{},
	); err != nil {
		panic(err)
	}
	return
}
