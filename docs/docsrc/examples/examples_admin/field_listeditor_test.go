package examples_admin

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"gorm.io/gorm"
)

// TestItem represents a single item in a list
type TestItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TestItems is a slice type that implements Value and Scan for JSONB handling
type TestItems []TestItem

func (t TestItems) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TestItems) Scan(value interface{}) error {
	if value == nil {
		*t = TestItems{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			*t = TestItems{}
			return nil
		}
		return json.Unmarshal([]byte(v), t)
	case []byte:
		if len(v) == 0 {
			*t = TestItems{}
			return nil
		}
		return json.Unmarshal(v, t)
	default:
		return errors.New("unsupported type for TestItems")
	}
}

// TestListContainer represents a container with list of items
type TestListContainer struct {
	gorm.Model
	Title string    `json:"title"`
	Items TestItems `gorm:"type:jsonb" json:"items"`
}

func TestListEditorAddRowBtnLabel(t *testing.T) {
	TestDB.AutoMigrate(&TestListContainer{})

	cases := []multipartestutils.TestCase{
		{
			Name:  "default add row button label",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createListEditorExample(presets.New(), TestDB, func(pb *presets.Builder, mb *presets.ModelBuilder) {
					// Use default AddRow label
					fb := pb.NewFieldsBuilder(presets.WRITE).Model(&TestItem{}).Only("Name", "Description")
					mb.Editing().Field("Items").Nested(fb)
				})
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &TestListContainer{
					Model: gorm.Model{ID: 1},
					Title: "Test Container",
					Items: TestItems{
						{Name: "Item 1", Description: "Description 1"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/test-list-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Item"}, // Default label
		},
		{
			Name:  "custom add row button label",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createListEditorExample(presets.New(), TestDB, func(pb *presets.Builder, mb *presets.ModelBuilder) {
					// Use custom AddRow label
					fb := pb.NewFieldsBuilder(presets.WRITE).Model(&TestItem{}).Only("Name", "Description")
					mb.Editing().Field("Items").Nested(fb, &presets.AddRowBtnLabel{
						LabelFunc: func(msgr *presets.Messages) string {
							return "Add New Item"
						},
					})
				})
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &TestListContainer{
					Model: gorm.Model{ID: 2},
					Title: "Test Container 2",
					Items: TestItems{
						{Name: "Item 2", Description: "Description 2"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/test-list-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "2").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add New Item"}, // Custom label
		},
		{
			Name:  "i18n add row button label helper",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createListEditorExample(presets.New(), TestDB, func(pb *presets.Builder, mb *presets.ModelBuilder) {
					// Use i18n helper function
					fb := pb.NewFieldsBuilder(presets.WRITE).Model(&TestItem{}).Only("Name", "Description")
					mb.Editing().Field("Items").Nested(fb, presets.AddRowLabelI18n(func(msgr *presets.Messages) string {
						return "Add More Items"
					}))
				})
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &TestListContainer{
					Model: gorm.Model{ID: 3},
					Title: "Test Container 3",
					Items: TestItems{
						{Name: "Item 3", Description: "Description 3"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/test-list-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "3").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add More Items"}, // I18n helper label
		},
		{
			Name:  "add row button label verification",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createListEditorExample(presets.New(), TestDB, func(pb *presets.Builder, mb *presets.ModelBuilder) {
					fb := pb.NewFieldsBuilder(presets.WRITE).Model(&TestItem{}).Only("Name", "Description")
					mb.Editing().Field("Items").Nested(fb, &presets.AddRowBtnLabel{
						LabelFunc: func(msgr *presets.Messages) string {
							return "Add Test Item"
						},
					})
				})
			},
			ReqFunc: func() *http.Request {
				// Seed data with empty list to show the add button
				container := &TestListContainer{
					Model: gorm.Model{ID: 4},
					Title: "Test Container 4",
					Items: TestItems{},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/test-list-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "4").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Test Item"}, // Custom button label
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}

func createListEditorExample(b *presets.Builder, db *gorm.DB, customize func(pb *presets.Builder, mb *presets.ModelBuilder)) http.Handler {
	db.AutoMigrate(&TestListContainer{})

	// Setup the project name, ORM and Homepage
	b.DataOperator(gorm2op.DataOperator(db))

	// Register TestListContainer into the builder
	mb := b.Model(&TestListContainer{})
	mb.Listing("ID", "Title")
	mb.Editing("Title", "Items")

	if customize != nil {
		customize(b, mb)
	}

	return b
}
