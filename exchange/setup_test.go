package exchange_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

var tables = []interface{}{
	&TestExchangeModel{},
	&TestExchangeCompositePrimaryKeyModel{},
	&Phone{},
	&Intro{},
	&ExtraIntro{},
	&Camera{},
	&ShoppingSite{},
}

type TestExchangeModel struct {
	ID       uint `gorm:"primarykey"`
	Name     string
	Age      *int
	Birth    *time.Time
	Appender string
}

type TestExchangeCompositePrimaryKeyModel struct {
	ID       uint   `gorm:"primarykey"`
	Name     string `gorm:"primarykey"`
	Age      *int
	Appender string
}

func TestMain(m *testing.M) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	db.Logger = db.Logger.LogMode(logger.Info)

	migrateTables()

	s := m.Run()
	// dropTables()
	os.Exit(s)
}

func migrateTables() {
	if err := db.AutoMigrate(tables...); err != nil {
		panic(err)
	}
}

func dropTables() {
	var err error
	for _, m := range tables {
		stmt := &gorm.Statement{DB: db}
		stmt.Parse(m)
		err = db.Exec(fmt.Sprintf("drop table %s", stmt.Schema.Table)).Error
		if err != nil {
			panic(err)
		}
	}

	err = db.Exec(fmt.Sprintf("drop table phone_selling_shopping_site")).Error
	if err != nil {
		panic(err)
	}
}

func initTables() {
	dropTables()
	migrateTables()
}

func ptrInt(v int) *int {
	return &v
}

func ptrTime(v time.Time) *time.Time {
	return &v
}
