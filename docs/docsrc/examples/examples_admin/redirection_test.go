package examples_admin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/redirection"
)

var redirectionData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.redirections (id, created_at, updated_at, deleted_at, source, target) VALUES (1, '2025-02-06 02:29:02.231371 +00:00', '2025-02-06 02:29:02.231371 +00:00', null, '/international/international/index3.html', 'https://www.xxx.com');

`, []string{"redirections"}))

func TestRedirection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	RedirectionExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/redirections", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"/international/international/index3.html", "https://www.xxx.com"},
		},
		{
			Name:  "Upload File Csv Format Error",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/redirections").
					AddReader("NewFiles", "test.csv", bytes.NewReader([]byte(`source,target
https://cdn.qor5.theplant-dev.com/international/index4.html,/international/index.html
/index4.html,international/international/index.html`))).
					EventFunc(redirection.UploadFileEvent).BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`RedirectionNotifyErrorMsg`, `Source Invalid Format`},
		},
		{
			Name:  "Upload File Source Is Duplicated",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/redirections").
					AddReader("NewFiles", "test.csv", bytes.NewReader([]byte(`source,target
/international/index4.html,/international/index.html
/international/index4.html,/international/index.html`))).
					EventFunc(redirection.UploadFileEvent).BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`RedirectionNotifyErrorMsg`, `Source Is Duplicated`},
		},
		{
			Name:  "Upload File Target Is Unreachable",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/redirections").
					AddReader("NewFiles", "test.csv", bytes.NewReader([]byte(`source,target
/international/index4.html,https://wwwwwwww/
/international/index5.html,https://wwwwwwww/`))).
					EventFunc(redirection.UploadFileEvent).BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`RedirectionNotifyErrorMsg`, `Target Is Unreachable`},
		},

		{
			Name:  "Upload File No File",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/redirections").
					EventFunc(redirection.UploadFileEvent).BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`RedirectionNotifyErrorMsg`, `File Upload Failed`},
		},

		{
			Name:  "Upload File No Records",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/redirections").
					AddReader("NewFiles", "test.csv", bytes.NewReader([]byte(`source,target`))).
					EventFunc(redirection.UploadFileEvent).BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`success`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
