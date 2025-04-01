package integration_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	h "github.com/theplant/htmlgo"
)

// TestWrapSaveFunc tests the WrapSaveFunc method of SectionBuilder
func TestWrapSaveFunc(t *testing.T) {
	db := TestDB
	// Create required tables
	err := db.AutoMigrate(&ParameterSetting{})
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Get sql.DB from gorm.DB for test fixtures
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Create test data
	settingData := gofixtures.Data(gofixtures.Sql(`
	INSERT INTO parameter_setting (id, created_at, updated_at, deleted_at, parameter_id, display_name, description, visible_to_non_developer, condition_id, form_setting) 
	VALUES (1, '2023-01-01 00:00:00', '2023-01-01 00:00:00', NULL, 0, 'originalName', '', 0, NULL, '[{"Path":"/path1","ValType":"STRING","Description":"desc1","DisplayName":"name1"}]');
	`, []string{"parameter_setting"}))
	settingData.TruncatePut(sqlDB)

	// Create preset builder with GORM data operator
	b := presets.New()
	b.DataOperator(gorm2op.DataOperator(db))
	b.URIPrefix("/ps")

	// Register model
	mb := b.Model(&ParameterSetting{})

	// Configure detail builder
	detail := mb.Detailing()

	// Define test cases
	type TestCase struct {
		Name         string
		SectionSetup func(section *presets.SectionBuilder)
		Request      func(db *sql.DB) *http.Request
		ExpectedName string
	}

	testCases := []TestCase{
		{
			Name: "DefaultSaveFunc",
			SectionSetup: func(section *presets.SectionBuilder) {
				// Default behavior without WrapSaveFunc
				section.Editing("DisplayName")
			},
			Request: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_save_Detail").
					Query(presets.ParamID, "1").
					AddField("DisplayName", "newName").
					BuildEventFuncRequest()
			},
			ExpectedName: "newName", // Default behavior - name changed directly
		},
		{
			Name: "WrappedSaveFunc",
			SectionSetup: func(section *presets.SectionBuilder) {
				// Add WrapSaveFunc that modifies the DisplayName
				section.WrapSaveFunc(func(original presets.SaveFunc) presets.SaveFunc {
					return func(obj interface{}, id string, ctx *web.EventContext) error {
						// Modify the object before saving
						if ps, ok := obj.(*ParameterSetting); ok {
							ps.DisplayName = "Wrapped_" + ps.DisplayName
						}
						// Call the original SaveFunc
						return original(obj, id, ctx)
					}
				}).Editing("DisplayName")
			},
			Request: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_save_Detail").
					Query(presets.ParamID, "1").
					AddField("DisplayName", "newName").
					BuildEventFuncRequest()
			},
			ExpectedName: "Wrapped_newName", // Wrapped behavior - prefix added to name
		},
	}

	// Create a section using NewSectionBuilder
	detailSection := presets.NewSectionBuilder(mb, "Detail")

	// Configure section display
	detailSection.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div().Text(fmt.Sprintf("Detail Section for %v", obj))
	})

	// Register section with detail
	detail.Section(detailSection)

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Configure section for this test case
			tc.SectionSetup(detailSection)

			// Set up HTTP handler and recorder
			w := httptest.NewRecorder()
			req := tc.Request(sqlDB)

			// Process the request
			b.ServeHTTP(w, req)

			// Verify results
			var ps ParameterSetting
			if err := db.First(&ps).Error; err != nil {
				t.Fatalf("Failed to query result: %v", err)
			}

			// Check if DisplayName was modified as expected
			if ps.DisplayName != tc.ExpectedName {
				t.Errorf("Expected DisplayName to be %q, got %q", tc.ExpectedName, ps.DisplayName)
			}
		})
	}
}
