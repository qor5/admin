package admin

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/role"
	"github.com/qor5/x/perm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// ConnectDB is used to connect DB with DB_PARAMS env.
func ConnectDB() *gorm.DB {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	if err = migrateModels(db); err != nil {
		panic(err)
	}

	return db
}

// ConnectRDS is used to connect RDS with IAM.
func ConnectRDS() *gorm.DB {
	var (
		dbName string = os.Getenv("AWS_DB_NAME")
		dbUser string = os.Getenv("AWS_DB_USER")
		dbHost string = os.Getenv("AWS_DB_HOST")
		port   string = os.Getenv("AWS_DB_PORT")
		region string = os.Getenv("AWS_REGION")
	)

	// Connect DB with DB_PARAMS env if AWS env is null.
	if dbName == "" {
		return ConnectDB()
	}

	dbPort, _ := strconv.Atoi(port)

	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		dbHost, dbPort, dbUser, authenticationToken, dbName,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)

	if err = migrateModels(db); err != nil {
		panic(err)
	}

	return db
}

func migrateModels(db *gorm.DB) (err error) {
	return db.AutoMigrate(
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
	)
}
