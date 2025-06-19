package examples_admin

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"gorm.io/gorm"
)

// ExampleItem represents a single item in a list for demonstration
type ExampleItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ExampleItems is a slice type that implements Value and Scan for JSONB handling
type ExampleItems []ExampleItem

func (t ExampleItems) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *ExampleItems) Scan(value interface{}) error {
	if value == nil {
		*t = ExampleItems{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	default:
		return errors.New("cannot scan into ExampleItems")
	}
}

// ExampleContainer represents a container with list of items for demonstration
type ExampleContainer struct {
	gorm.Model
	Title string       `json:"title"`
	Items ExampleItems `gorm:"type:jsonb" json:"items"`
}

// ListEditorAddRowBtnLabelExample demonstrates how to customize the "Add Row" button label
// in list editor components using different approaches
func ListEditorAddRowBtnLabelExample(b *presets.Builder, db *gorm.DB) http.Handler {
	db.AutoMigrate(&ExampleContainer{})

	// Setup the project name, ORM and Homepage
	b.DataOperator(gorm2op.DataOperator(db))

	// Register ExampleContainer into the builder
	mb := b.Model(&ExampleContainer{})
	mb.Listing("ID", "Title")

	// Configure editing fields with different AddRowBtnLabel examples
	eb := mb.Editing("Title", "Items")

	// Example 1: Using custom label with LabelFunc
	fb := b.NewFieldsBuilder(presets.WRITE).Model(&ExampleItem{}).Only("Name", "Description")
	eb.Field("Items").Nested(fb, &presets.AddRowBtnLabel{
		LabelFunc: func(msgr *presets.Messages) string {
			return "Add New Item"
		},
	})

	// Seed some sample data
	seedData := []ExampleContainer{
		{
			Model: gorm.Model{ID: 1},
			Title: "Sample Container 1",
			Items: ExampleItems{
				{Name: "Item 1", Description: "First sample item"},
				{Name: "Item 2", Description: "Second sample item"},
			},
		},
		{
			Model: gorm.Model{ID: 2},
			Title: "Sample Container 2",
			Items: ExampleItems{
				{Name: "Item A", Description: "Sample item A"},
			},
		},
		{
			Model: gorm.Model{ID: 3},
			Title: "Empty Container",
			Items: ExampleItems{},
		},
	}

	for _, data := range seedData {
		var existing ExampleContainer
		if err := db.First(&existing, data.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				db.Create(&data)
			}
		}
	}

	return b
}
