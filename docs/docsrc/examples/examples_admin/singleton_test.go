package examples_admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestSingletonExample(t *testing.T) {
	pb := presets.New()
	singletonExample(pb, TestDB, nil)

	cases := []multipartestutils.TestCase{
		{
			Name:  "index",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-singleten-products", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Editing WithSingletenProduct", "Title", `"Title":""`, `.eventFunc("presets_Update").queries({"id":["1"]})`},
		},
		{
			Name:  "update",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/with-singleten-products?__execute_event__=presets_Update&id=1").
					AddField("Title", `Hello World`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`Successfully Updated`},
		},
		{
			Name:  "update with force error returned",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				singletonExample(pb, TestDB, func(mb *presets.ModelBuilder) {
					mb.Editing().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
						return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
							return errors.New("force error")
							// err = in(obj, id, ctx)
							// return err
						}
					})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/with-singleten-products?__execute_event__=presets_Update&id=1").
					AddField("Title", `Hello World`).
					BuildEventFuncRequest()
			},

			ExpectPortalUpdate0ContainsInOrder: []string{"force error", "Title", `"Title":"Hello World"`, `.eventFunc("presets_Update").queries({"id":["1"]})`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}
