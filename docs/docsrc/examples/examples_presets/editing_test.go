package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var companyData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.companies (id, name) VALUES (12, 'terry_company');
`, []string{"companies"}))

func TestPresetsEditingValidate(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingValidate(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "editing create",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"timeout='2000'", "name must not be empty"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingSetter(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSetter(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "default field setterFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name must not be empty"},
		},
		{
			Name:  "wrap setter",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "system").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`You can not use \"system\" as name`},
		},
		{
			Name:  "setter return global error",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "global").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"You can not use global as name"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingCustomizationDescription(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingCustomizationDescription(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Customer Name`, `Customer Email`, `Description`, `vx-tiptap-editor`},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("Body", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					AddField("Body_richeditor_medialibrary.Values", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`},
		},
		{
			Name:  "do new without avatar",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
				mediaBuilder := media.New(TestDB).AutoMigrate()
				defer pb.Use(mediaBuilder)

				mb, _, _, _ := PresetsEditingCustomizationDescription(pb, TestDB)
				mb.Editing().Field("Avatar").WithContextValue(media.MediaBoxConfig, &media_library.MediaBoxConfig{})
				mb.Editing().WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
					return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
						err = in(obj, ctx)
						customer := obj.(*Customer)
						if customer.Avatar.FileName == "" || customer.Avatar.Url == "" {
							err.FieldError("Avatar", "avatar must not be empty")
						}
						return
					}
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("Body", "").
					AddField("Avatar.Values", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					AddField("Body_richeditor_medialibrary.Values", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`, `avatar must not be empty`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingTiptap(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingTiptap(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Customer Name`, `Customer Email`, `Description`, `tiptap`},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingSection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSection(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`section1`, "Name", "Create"},
			ExpectPortalUpdate0NotContains:     []string{"Save", "Edit"},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("section1.Name", "Terry").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Terry"},
		},
		{
			Name:  "detailing section use editing validator",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=section_save_section1").
					AddField("section1.Name", "terryterryterry").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`too long name`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
