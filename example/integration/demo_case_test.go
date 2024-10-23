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
			ExpectPageBodyContainsInOrder: []string{"vx-field", "&#34;121231321&amp;&amp;&#34;", "vx-select", "vx-checkbox"},
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
					AddField("FieldSection.FieldData.Textarea", "456").
					AddField("FieldSection.FieldData.TextValidate", "12345").
					AddField("FieldSection.FieldData.TextareaValidate", "1234567890").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := admin.DemoCase{}
				TestDB.Order("id desc").First(&m, 1)
				if m.FieldData.Text != "123" {
					t.Fatalf("Update Demo Case Field Failed: %v", m.FieldData)
				}
				return
			},
		},
		{
			Name:  "Demo Case Field TextValidate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldSection").
					Query(presets.ParamID, "1").
					AddField("FieldSection.FieldData.Text", "123").
					AddField("FieldSection.FieldData.Textarea", "456").
					AddField("FieldSection.FieldData.TextValidate", "1234").
					AddField("FieldSection.FieldData.TextareaValidate", "1234567890").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"input more than 5 chars"},
		},
		{
			Name:  "Demo Case Field TextareaValidate",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoCaseData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo-cases/1").
					EventFunc("section_save_FieldSection").
					Query(presets.ParamID, "1").
					AddField("FieldSection.FieldData.Text", "123").
					AddField("FieldSection.FieldData.Textarea", "456").
					AddField("FieldSection.FieldData.TextValidate", "12345").
					AddField("FieldSection.FieldData.TextareaValidate", "123456789").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"input more than 10 chars"},
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
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
