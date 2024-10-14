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
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Edit post body",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=presets_Detailing_Field_Save&id=1_2023-01-05-v01").
					Query("section", "Detail").
					AddField("Detail.Title", "Demo").
					AddField("Detail.Body", "<p>test edit</p>").
					AddField("Detail.HeroImage.Description", "").
					AddField("Detail.HeroImage.Values", "{\"ID\":1,\"Url\":\"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/1/file.jpeg\",\"VideoLink\":\"\",\"FileName\":\"demo image.jpeg\",\"Description\":\"\",\"FileSizes\":{\"@qor_preview\":8917,\"default\":326350,\"main\":94913,\"og\":123973,\"original\":326350,\"thumb\":21199,\"twitter-large\":117784,\"twitter-small\":77615},\"Width\":750,\"Height\":1000}").
					AddField("Detail.Body_richeditor_medialibrary.Values", "").
					AddField("Detail.BodyImage.Values", "").
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
					EventFunc(actions.DoEditDetailingField).
					Query(presets.ParamID, "1_2023-01-05-v01").
					Query("section", "Detail").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Title`, `vx-field`, "Hero Image", "Choose File", "Body", "vx-tiptap-editor"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
