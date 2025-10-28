package integration_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/theplant/gofixtures"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/testingutils"
	"gorm.io/gorm"
)

type ParameterSetting struct {
	gorm.Model
	ParaMeterSettingDetail
	// This is where we setup the form. It should be a collection edit, each item contains 2 attributes
	// [ {"path": "path to the value", "valType": "string"}, {"path": "path to the value", "valType": "boolean"}]
	FormSetting ParameterFieldSettingArray `gorm:"type:text;" sql:"type:text"`
}

func (*ParameterSetting) TableName() string {
	return "parameter_setting"
}

type (
	ParameterFieldSettingArray []*ParameterFieldSetting
	ParameterConditionIDArray  []uint
)

func (parameterFieldSettingArray ParameterFieldSettingArray) Value() (driver.Value, error) {
	json, err := json.Marshal(parameterFieldSettingArray)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (parameterFieldSettingArray *ParameterFieldSettingArray) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		return json.Unmarshal(values, &parameterFieldSettingArray)
	case string:
		return parameterFieldSettingArray.Scan([]byte(values))
	}
	return nil
}

func (ParameterConditionIDArray ParameterConditionIDArray) Value() (driver.Value, error) {
	json, err := json.Marshal(ParameterConditionIDArray)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (ParameterConditionIDArray *ParameterConditionIDArray) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		return json.Unmarshal(values, &ParameterConditionIDArray)
	case string:
		return ParameterConditionIDArray.Scan([]byte(values))
	}
	return nil
}

type ParaMeterSettingDetail struct {
	ParameterID           uint
	DisplayName           string
	Description           string
	VisibleToNonDeveloper bool                      `gorm:"default:FALSE"` // Control if this parameter is visible to content operators
	ConditionID           ParameterConditionIDArray `gorm:"type:text;" sql:"type:text"`
}

type ParameterFieldSetting struct {
	Path        string // "/biscuits/0/Name"
	ValType     string // It should be a selector.
	Description string
	DisplayName string
	// Options     [][]string // When we want the val to be displayed as a selector. fill this.
}

type Case struct {
	Name      string
	FieldName string
	Field     func(b *presets.DetailingBuilder)
	ReqFunc   func(db *sql.DB) *http.Request
	Want      ParameterSetting
}

var settingData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO parameter_setting (id, created_at, updated_at, deleted_at, parameter_id, display_name, description, visible_to_non_developer, condition_id, form_setting) VALUES (1, '0001-01-01 00:00:00.000000 +00:00', '2024-04-15 02:00:49.123472 +00:00', null, 0, 'oldName', '', false, null, '[{"Path":"/path1","ValType":"STRING","Description":"desc1","DisplayName":"name1"}]');
			`, []string{"parameter_setting"}))

func TestDetailFieldBuilder(t *testing.T) {
	Cases := []Case{
		{
			Name:      "DetailFieldSave",
			FieldName: "Detail",
			ReqFunc: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_save_Detail").
					Query(presets.ParamID, "1").
					AddField("DisplayName", "newName").
					BuildEventFuncRequest()
			},
			Want: ParameterSetting{
				Model: gorm.Model{
					ID: 1,
				},
				ParaMeterSettingDetail: ParaMeterSettingDetail{
					DisplayName: "oldName",
				},
				FormSetting: ParameterFieldSettingArray{{
					Path:        "/path1",
					ValType:     "STRING",
					Description: "desc1",
					DisplayName: "name1",
				}},
			},
		},
		{
			Name:      "DetailListFieldCreate",
			FieldName: "FormSetting",
			ReqFunc: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_create_FormSetting").
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
			},
			Want: ParameterSetting{
				Model: gorm.Model{
					ID: 1,
				},
				ParaMeterSettingDetail: ParaMeterSettingDetail{
					DisplayName: "oldName",
				},
				FormSetting: []*ParameterFieldSetting{
					{
						Path:        "/path1",
						ValType:     "STRING",
						Description: "desc1",
						DisplayName: "name1",
					},
				},
			},
		},
		{
			Name:      "DetailListFieldUpdate",
			FieldName: "FormSetting",
			ReqFunc: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_save_FormSetting").
					Query(presets.ParamID, "1").
					Query("sectionListSaveBtn_FormSetting", "0").
					AddField("FormSetting[0].Path", "/newPath").
					AddField("FormSetting[0].DisplayName", "newName").
					AddField("FormSetting[0].Description", "newDesc").
					AddField("FormSetting[0].ValType", "/IMAGE").
					AddField("FormSetting[1].ValType", "/NUMBER").
					AddField("FormSetting[1].DisplayName", "otherName").
					AddField("FormSetting[1].Description", "otherDesc").
					AddField("FormSetting[1].Path", "/otherPath").
					BuildEventFuncRequest()
			},
			Want: ParameterSetting{
				Model: gorm.Model{
					ID: 1,
				},
				ParaMeterSettingDetail: ParaMeterSettingDetail{
					DisplayName: "oldName",
				},
				FormSetting: []*ParameterFieldSetting{{
					Path:        "/newPath",
					ValType:     "/IMAGE",
					Description: "newDesc",
					DisplayName: "newName",
				}},
			},
		},
		{
			Name:      "DetailListFieldDelete",
			FieldName: "FormSetting",
			ReqFunc: func(db *sql.DB) *http.Request {
				settingData.TruncatePut(db)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/ps/parameter-settings").
					EventFunc("section_delete_FormSetting").
					Query(presets.ParamID, "1").
					Query("sectionListDeleteBtn_FormSetting", "0").
					AddField("FormSetting[0].Path", "/newPath").
					AddField("FormSetting[0].DisplayName", "newName").
					AddField("FormSetting[0].Description", "newDesc").
					AddField("FormSetting[0].ValType", "/IMAGE").
					BuildEventFuncRequest()
			},
			Want: ParameterSetting{
				Model: gorm.Model{
					ID: 1,
				},
				ParaMeterSettingDetail: ParaMeterSettingDetail{
					DisplayName: "oldName",
				},
				FormSetting: []*ParameterFieldSetting{},
			},
		},
	}

	db := TestDB
	db.AutoMigrate(&ParameterSetting{})
	b := presets.New().URIPrefix("/ps")
	b.DataOperator(gorm2op.DataOperator(db))

	cust := b.Model(&ParameterSetting{})

	detail := cust.Detailing("ParameterID", "Detail", "FormSetting").Drawer(true)
	detailSection := presets.NewSectionBuilder(cust, "Detail").
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			ps := obj.(*ParameterSetting)
			return h.Div(h.Text(ps.DisplayName))
		}).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			ps := obj.(*ParameterSetting)
			return h.Div(
				v.VTextField().
					Attr(web.VField(fmt.Sprintf("%s.DisplayName", field.FormKey), ps.DisplayName)...),
			)
		})
	formSetting := presets.NewSectionBuilder(cust, "FormSetting").
		IsList(&ParameterFieldSetting{}).
		Editing("DisplayName", "Description", "Path", "ValType").
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			ps := obj.(*ParameterFieldSetting)
			return h.Div(
				h.Span(ps.DisplayName),
				h.Span(ps.Description),
				h.Span(ps.Path),
				h.Span(ps.ValType),
			)
		}).
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			ps := obj.(*ParameterFieldSetting)
			div := h.Div(
				v.VTextField().
					Variant("outlined").Density("compact").Placeholder("DisplayName").
					Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "DisPlayName"), ps.DisplayName)...),
				v.VTextField().
					Variant("outlined").Density("compact").Placeholder("Description").
					Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "Description"), ps.Description)...),
				v.VTextField().
					Variant("outlined").Density("compact").Placeholder("Path").
					Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "Path"), ps.Path)...),
				v.VTextField().
					Variant("outlined").Density("compact").Placeholder("ValType").
					Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "ValType"), ps.ValType)...),
			)
			return div
		})
	detail.Section(detailSection, formSetting)
	for _, c := range Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			dbraw, _ := db.DB()

			r := c.ReqFunc(dbraw)

			b.ServeHTTP(w, r)

			var ps *ParameterSetting
			err := db.First(&ps).Error
			if err != nil {
				t.Error(err)
			}
			if diff := testingutils.PrettyJsonDiff(ps.ParaMeterSettingDetail, c.Want.ParaMeterSettingDetail); diff != "" {
				t.Error(diff)
			}
			if diff := testingutils.PrettyJsonDiff(ps.FormSetting, c.Want.FormSetting); diff != "" {
				t.Error(diff)
			}
		})
	}
}

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
	VALUES (1, '2023-01-01 00:00:00', '2023-01-01 00:00:00', NULL, 0, 'originalName', '', FALSE, NULL, '[{"Path":"/path1","ValType":"STRING","Description":"desc1","DisplayName":"name1"}]');
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
