package dynamic_test

import (
	"testing"

	dynamicstruct "github.com/ompluscator/dynamic-struct"
	"github.com/sunfmin/reflectutils"
	"github.com/theplant/testingutils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNew(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	itemsType := dynamicstruct.NewStruct().
		AddField("ID", uint(0), `gorm:"primaryKey" json:"-"`).
		AddField("Name", "", `json:"name"`).
		AddField("TestTableID", "", `json:"-"`).
		Build()
	items := itemsType.NewSliceOfStructs()
	instance := dynamicstruct.NewStruct().
		AddField("Integer", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("SliceOfItems", items, `gorm:"foreignKey:TestTableID"`).
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	reflectutils.Set(instance, "Integer", 1)
	if reflectutils.MustGet(instance, "Integer").(int) != 1 {
		t.Error("Integer field is not set")
	}

	items2 := itemsType.NewSliceOfStructs()
	reflectutils.Set(items2, "[0].Name", "test1")
	reflectutils.Set(items2, "[1].Name", "test2")
	reflectutils.Set(instance, "SliceOfItems", items2)
	diff := testingutils.PrettyJsonDiff(reflectutils.MustGet(instance, "SliceOfItems"),
		[]struct {
			Name string `json:"name"`
		}{
			{Name: "test1"},
			{Name: "test2"},
		})

	if len(diff) > 0 {
		t.Error(diff)
	}

	err = db.Table("test_tables").Create(instance).Error
	if err != nil {
		panic(err)
	}
}
