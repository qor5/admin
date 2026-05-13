package examples_admin

import (
	"context"
	"database/sql"
	"testing"

	"github.com/qor5/x/v3/gormx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	TestDB *gorm.DB
	SqlDB  *sql.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testSuite := gormx.MustStartTestSuite(ctx)
	defer testSuite.Stop(ctx)
	TestDB = testSuite.DB()
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	SqlDB, _ = TestDB.DB()
	m.Run()
}
