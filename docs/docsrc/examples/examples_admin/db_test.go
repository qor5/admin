package examples_admin

import (
	"database/sql"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/qor5/x/v3/gormx"
	"gorm.io/driver/postgres"
	"context"
)

var (
	TestDB *gorm.DB
	SqlDB  *sql.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := gormx.OpenContainer(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()
	TestDB, err = gorm.Open(postgres.Open(pgContainer.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	SqlDB, _ = TestDB.DB()
	m.Run()
}
