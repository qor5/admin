package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var demoCaseData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.demo_cases (id, created_at, updated_at, deleted_at, name,field_data) VALUES (1, '2024-10-10 03:18:50.316417 +00:00', '2024-10-10 03:18:50.316417 +00:00', null, '12313','{"Text":"121231321\u0026\u0026","Textarea":"1231","TextValidate":"21312","TextareaValidate":"1😋11231"}');
`, []string{`demo_cases`}))

func TestDemoCase(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Demo Case",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/demo-cases", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Name", "12313"},
		},
		{
			Name:  "Create Demo Case",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases").
					EventFunc(actions.Update).
					AddField("Name", "test").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m)
				if m.Name != "test" {
					t.Fatalf("Create Demo Case Failed: %v", m)
				}
				return
			},
		},
		{
			Name:  "Demo Case Detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{
				"vx-field",
				"&#34;121231321&amp;&amp;&#34;",
				"vx-field(type textarea)",
				"vx-field(type password)",
				"vx-field(type number)",
				"vx-select",
				"vx-checkbox",
				"vx-datepicker",
				"vx-dialog",
				"vx-avatar",
			},
		},
		{
			Name:  "Demo Case Field Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldSection").
					Query(presets.ParamID, "1").
					AddField("FieldSection.FieldData.Text", "123").
					AddField("FieldSection.FieldData.TextValidate", "12345").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if m.FieldData.TextValidate != "12345" {
					t.Fatalf("Update Demo Case Field Failed: %v", m.FieldData)
				}
				return
			},
		},
		{
			Name:  "Demo Case Field Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldSection").
					Query(presets.ParamID, "1").
					AddField("FieldSection.FieldData.Text", "123").
					AddField("FieldSection.FieldData.TextValidate", "1234").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"input more than 5 chars"},
		},
		{
			Name:  "Demo Case FieldTextarea Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldTextareaSection").
					Query(presets.ParamID, "1").
					AddField("FieldTextareaSection.FieldTextareaData.Textarea", "456").
					AddField("FieldTextareaSection.FieldTextareaData.TextareaValidate", "1234567890").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if m.FieldTextareaData.TextareaValidate != "1234567890" {
					t.Fatalf("Update Demo Case Field Failed: %v", m.FieldTextareaData)
				}
				return
			},
		},
		{
			Name:  "Demo Case FieldTextarea Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldTextareaSection").
					Query(presets.ParamID, "1").
					AddField("FieldTextareaSection.FieldTextData.Textarea", "1234").
					AddField("FieldTextareaSection.FieldTextData.TextareaValidate", "1234").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"input more than 10 chars"},
		},
		{
			Name:  "Demo Case FieldPassword Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldPasswordSection").
					Query(presets.ParamID, "1").
					AddField("FieldPasswordSection.FieldPasswordData.Password", "12345").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if m.FieldPasswordData.Password != "12345" {
					t.Fatalf("Update Demo Case Field Failed: %v", m.FieldPasswordData)
				}
				return
			},
		},
		{
			Name:  "Demo Case FieldPassword Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldPasswordSection").
					Query(presets.ParamID, "1").
					AddField("FieldPasswordSection.FieldPasswordData.Password", "1234").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"password more than 5 chars"},
		},
		{
			Name:  "Demo Case FieldNumber Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldNumberSection").
					Query(presets.ParamID, "1").
					AddField("FieldNumberSection.FieldNumberData.Number", "0").
					AddField("FieldNumberSection.FieldNumberData.NumberValidate", "20").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if m.FieldNumberData.NumberValidate != 20 {
					t.Fatalf("Update Demo Case Field Failed: %v", m.FieldNumberData)
				}
				return
			},
		},
		{
			Name:  "Demo Case FieldNumber Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldNumberSection").
					Query(presets.ParamID, "1").
					AddField("FieldNumberSection.FieldNumberData.Number", "20").
					AddField("FieldNumberSection.FieldNumberData.NumberValidate", "0").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"input greater than 0"},
		},
		{
			Name:  "Demo Case Select Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_SelectSection").
					Query(presets.ParamID, "1").
					AddField("SelectSection.SelectData.AutoComplete[0]", "1").
					AddField("SelectSection.SelectData.AutoComplete[1]", "2").
					AddField("SelectSection.SelectData.NormalSelect", "3").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if len(m.SelectData.AutoComplete) == 0 {
					t.Fatalf("Update Demo Case Field Failed: %v", m.SelectData)
					return
				}
				if m.SelectData.AutoComplete[0] != 1 {
					t.Fatalf("Update Demo Case Field Failed: %v", m.SelectData)
				}
				if m.SelectData.AutoComplete[1] != 2 {
					t.Fatalf("Update Demo Case Field Failed: %v", m.SelectData)
				}
				if m.SelectData.NormalSelect != 3 {
					t.Fatalf("Update Demo Case Field Failed: %v", m.SelectData)
				}
				return
			},
		},
		{
			Name:  "Demo Case Select Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_SelectSection").
					Query(presets.ParamID, "1").
					AddField("SelectSection.SelectData.AutoComplete[0]", "1").
					AddField("SelectSection.SelectData.NormalSelect", "8").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"select more than 1 item", "can`t select Trevor"},
		},
		{
			Name:  "Demo Case CheckBox Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_CheckboxSection").
					Query(presets.ParamID, "1").
					AddField("CheckboxSection.CheckboxData.Checkbox", "true").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if !m.CheckboxData.Checkbox {
					t.Fatalf("Update Demo Case Field Failed: %v", m.CheckboxData)
				}
				return
			},
		},
		{
			Name:  "Demo Case DatePicker Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_DatepickerSection").
					Query(presets.ParamID, "1").
					AddField("DatepickerSection.DatepickerData.Date", "0").
					AddField("DatepickerSection.DatepickerData.DateTime", "0").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Date is required", "DateTime is required", "End later than Start"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
