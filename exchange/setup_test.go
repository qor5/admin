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
	&Camera{},
	&ShoppingSite{},
}

type TestExchangeModel struct {
	ID    uint `gorm:"primarykey"`
	Name  string
	Age   *int
	Birth *time.Time
}

type TestExchangeCompositePrimaryKeyModel struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"primarykey"`
	Age   *int
	Birth *time.Time
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

// type Variant struct {
// 	gorm.Model
// 	Code              string `sql:"not null;"`
// 	ProductID         uint
// 	ProductCode       string
// 	Images            string
// 	ExternalID        string
// 	Price             uint64
// 	SellingPrice      uint64
// 	VolumeCoefficient int64 // For convenience store pick store
// }

// func TestImportVariants(t *testing.T) {
// 	var err error
// 	err = db.AutoMigrate(&Variant{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	exchange.NewImporter(&Variant{}).Metas(
// 		exchange.NewMeta("Code").Header("JANコード"),
// 		exchange.NewMeta("ProductCode").Header("品番"),
// 		exchange.NewMeta("Price").Header("上代1"),
// 		exchange.NewMeta("SellingPrice").Header("上代2"),
// 		// exchange.NewMeta("ColorCode").Header("カラー"),
// 	)
// }
