package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/autosync"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestAutoSyncFrom(t *testing.T) {
	newLazyWrapperEditCompoSync := func(initialChecked autosync.InitialChecked) func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return autosync.NewLazyWrapComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) *autosync.Config {
			return &autosync.Config{
				SyncFromFromKey: strings.TrimSuffix(field.FormKey, "Slug"),
				InitialChecked:  initialChecked,
				CheckboxLabel:   "Auto Sync",
				SyncCall:        autosync.SyncCallSlug,
			}
		})
	}

	cases := []multipartestutils.TestCase{
		{
			Name:  "InitialCheckedAuto",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				return autoSyncExample(pb, TestDB, func(mb *presets.ModelBuilder) {
					mb.Editing().Field("TitleSlug").LazyWrapComponentFunc(newLazyWrapperEditCompoSync(autosync.InitialCheckedAuto))
				})
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-slug-products?__execute_event__=presets_New", http.NoBody)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Title", "Title Slug", `form["TitleSlug__AutoSync__"] = (plaid().slug(form["Title"]||"")) === form["TitleSlug"]`, "Auto Sync"},
		},
		{
			Name:  "InitialCheckedFalse",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				return autoSyncExample(pb, TestDB, func(mb *presets.ModelBuilder) {
					mb.Editing().Field("TitleSlug").LazyWrapComponentFunc(newLazyWrapperEditCompoSync(autosync.InitialCheckedFalse))
				})
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-slug-products?__execute_event__=presets_New", http.NoBody)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Title", "Title Slug", `form["TitleSlug__AutoSync__"] = false`, "Auto Sync"},
		},
		{
			Name:  "InitialCheckedTrue",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				return autoSyncExample(pb, TestDB, func(mb *presets.ModelBuilder) {
					mb.Editing().Field("TitleSlug").LazyWrapComponentFunc(newLazyWrapperEditCompoSync(autosync.InitialCheckedTrue))
				})
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-slug-products?__execute_event__=presets_New", http.NoBody)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Title", "Title Slug", `form["TitleSlug__AutoSync__"] = true`, "Auto Sync"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, presets.New())
		})
	}
}
