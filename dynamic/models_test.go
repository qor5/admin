package dynamic_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	dynamicstruct "github.com/ompluscator/dynamic-struct"
	"github.com/sunfmin/reflectutils"
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
	_ = items
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
	// reflectutils.Set(instance, "SliceOfItems", items2)
	// diff := testingutils.PrettyJsonDiff(reflectutils.MustGet(instance, "SliceOfItems"),
	// 	[]struct {
	// 		Name string `json:"name"`
	// 	}{
	// 		{Name: "test1"},
	// 		{Name: "test2"},
	// 	})
	//
	// if len(diff) > 0 {
	// 	t.Error(diff)
	// }

	err = db.Table("test_tables").AutoMigrate(instance)
	if err != nil {
		panic(err)
	}
	err = db.Table("test_tables").Create(instance).Error
	if err != nil {
		panic(err)
	}
}

func createStructFields(data map[string]interface{}) []reflect.StructField {
	var fields []reflect.StructField

	for key, value := range data {
		var fieldType reflect.Type

		switch v := value.(type) {
		case map[string]interface{}:
			nestedFields := createStructFields(v)
			fieldType = reflect.StructOf(nestedFields)
		default:
			fieldType = reflect.TypeOf(value)
		}

		field := reflect.StructField{
			Name: strings.Title(key),
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, key)),
		}

		fields = append(fields, field)
	}

	return fields
}

func populateStruct(sv reflect.Value, data map[string]interface{}) {
	for key, value := range data {
		field := sv.FieldByName(strings.Title(key))

		switch v := value.(type) {
		case map[string]interface{}:
			nestedStruct := reflect.New(field.Type()).Elem()
			populateStruct(nestedStruct, v)
			field.Set(nestedStruct)
		default:
			field.Set(reflect.ValueOf(value))
		}
	}
}

func TestJSON(t *testing.T) {
	jsonFromDB := `{"name":"John Doe","age":30,"address":{"street":"123 Main St","city":"Anytown","zip":"12345"},"hobbies":["reading","gaming","hiking"]}`

	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonFromDB), &jsonMap)
	if err != nil {
		panic(err)
	}

	fields := createStructFields(jsonMap)
	dynamicType := reflect.StructOf(fields)

	structValue := reflect.New(dynamicType).Elem()
	populateStruct(structValue, jsonMap)

	fmt.Printf("Dynamic Struct: %+v\n", structValue.Interface())
}
