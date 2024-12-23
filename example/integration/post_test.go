package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

func TestPost(t *testing.T) {
	h := admin.TestHandlerWorker(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Edit post body",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=section_save_Detail&id=1_2023-01-05-v01").
					Query("section", "Detail").
					AddField("Title", "Demo").
					AddField("Body", "<p>test edit</p>").
					AddField("HeroImage.Description", "").
					AddField("HeroImage.Values", "{\"ID\":1,\"Url\":\"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/1/file.jpeg\",\"VideoLink\":\"\",\"FileName\":\"demo image.jpeg\",\"Description\":\"\",\"FileSizes\":{\"@qor_preview\":8917,\"default\":326350,\"main\":94913,\"og\":123973,\"original\":326350,\"thumb\":21199,\"twitter-large\":117784,\"twitter-small\":77615},\"Width\":750,\"Height\":1000}").
					AddField("Body_richeditor_medialibrary.Values", "").
					AddField("BodyImage.Values", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`test edit`},
		},
		{
			Name:  "Edit post Detail View",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts").
					EventFunc("section_edit_Detail").
					Query(presets.ParamID, "1_2023-01-05-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Title`, `vx-field`, "Hero Image", "Choose File", "Body", "vx-tiptap-editor"},
		},
		{
			Name:  "Index Post View",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`Demo`, `vx-filter`, "Create Time"},
		},
		{
			Name:  "Post Validate Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts").
					Query(presets.ParamID, "1_2023-01-05-v01").
					EventFunc(actions.Validate).
					AddField("Title", "").
					AddField("TitleWithSlug", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{`Title Is Required`, `TitleWithSlug Is Required`},
		},
		{
			Name:  "Post Update Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts").
					Query(presets.ParamID, "1_2023-01-05-v01").
					EventFunc(actions.Update).
					AddField("Title", "").
					AddField("TitleWithSlug", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Title Is Required`, `TitleWithSlug Is Required`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
