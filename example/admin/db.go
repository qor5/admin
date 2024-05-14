package admin

import (
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/x/v3/perm"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db             *gorm.DB
	dbParamsString = osenv.Get("DB_PARAMS", "admin example database connection string", "")
)

func ConnectDB() *gorm.DB {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	if err = db.AutoMigrate(
		&models.Post{},
		&models.InputDemo{},
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
	); err != nil {
		panic(err)
	}
	return db
}
