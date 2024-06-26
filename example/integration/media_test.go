package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/role"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/admin"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var mediaTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (1, '2024-06-14 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', '{"FileName":"Snipaste_2024-06-14_10-06-12.png","Url":"/system/media_libraries/1/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
INSERT INTO public.media_libraries (id,user_id, created_at, updated_at, deleted_at, selected_type, file) VALUES (2, 888, '2024-06-15 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', '{"FileName":"test_search1.png","Url":"/system/media_libraries/2/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
INSERT INTO public.media_libraries (id,user_id, created_at, updated_at, deleted_at, selected_type, file) VALUES (3, 999,'2024-06-14 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', '{"FileName":"test_search2.png","Url":"/system/media_libraries/3/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
`, []string{"media_libraries"}))

func TestMedia(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Page Builder edit seo image",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/1_2024-06-19-v01_International?__execute_event__=mediaLibrary_ChooseFileEvent&field=SEO.OpenGraphImageFromMediaLibrary&media_id=1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Snipaste_2024-06-14_10-06-12.png"},
		},
		{
			Name:  "MediaLibrary Admin Role List",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-library").
					AddField("type", "all").
					AddField("order_by", "created_at_desc").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test_search1.png", "test_search2.png"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}

	cases = []TestCase{
		{
			Name:  "MediaLibrary Editor Role List",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-library").
					AddField("type", "all").
					AddField("order_by", "created_at_desc").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test_search2.png"},
			ExpectPageBodyNotContains:     []string{"test_search1.png"},
		},
	}

	h = admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 999},
		Roles: []role.Role{
			{
				Name: models.RoleEditor,
			},
		},
	})
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
