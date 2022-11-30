package admin

import (
	"os"

	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/role"
	"github.com/qor5/x/perm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func ConnectDB() *gorm.DB {
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
		&models.LoginSession{},
		&models.ListModel{},
		&role.Role{},
		&perm.DefaultDBPolicy{},
		&models.Customer{},
		&models.Address{},
		&models.Phone{},
		&models.MembershipCard{},
		&models.Product{},
		&models.Order{},
		&models.Category{},
		&models.MicrositeModel{},
	); err != nil {
		panic(err)
	}
	return db
}
