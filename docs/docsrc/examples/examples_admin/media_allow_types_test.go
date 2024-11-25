package examples_admin

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	media_oss "github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/theplant/gofixtures"
)

var allowTypesMediaData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, folder,file) VALUES (1, '2024-06-14 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', false,'{"FileName":"test_image.png"}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, folder,file) VALUES (2, '2024-06-15 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'video', false,'{"FileName":"test_video.mp4"}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, folder,file) VALUES (3, '2024-06-16 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'file', false,'{"FileName":"test_file.text"}');
`, []string{"media_libraries"}))

func TestMediaAllowTypesExample(t *testing.T) {
	dbr, _ := TestDB.DB()
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	MediaAllowTypesExample(pb, TestDB)
	cases := []TestCase{
		{
			Name:  "Allow Types List Media Library",
			Debug: true,
			ReqFunc: func() *http.Request {
				allowTypesMediaData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"image/*,video/*", "test_video.mp4", "test_image.png"},
			ExpectPageBodyNotContains:     []string{"test_file.text"},
		},
		{
			Name:  "Allow Types List Media Library Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				allowTypesMediaData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(media.OpenFileChooserEvent).
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test_video.mp4", "test_image.png"},
			ExpectPageBodyNotContains:     []string{"test_file.text"},
		},
		{
			Name:  "Media Library Tab upload image file",
			Debug: true,
			ReqFunc: func() *http.Request {
				allowTypesMediaData.TruncatePut(dbr)
				media_oss.Storage = filesystem.New("/tmp/media_test")
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.UploadFileEvent).
					Query(media.ParamField, "media").
					AddReader("NewFiles", "test2.png", bytes.NewReader([]byte("test upload file"))).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m)
				if m.File.FileName != "test2.png" {
					t.Fatalf("except filename not is test2.png but got %v", m.File.FileName)
					return
				}
				return
			},
		},
		{
			Name:  "Media Library Upload Text file",
			Debug: true,
			ReqFunc: func() *http.Request {
				allowTypesMediaData.TruncatePut(dbr)
				media_oss.Storage = filesystem.New("/tmp/media_test")
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.UploadFileEvent).
					Query(media.ParamField, "media").
					AddReader("NewFiles", "test2.txt", bytes.NewReader([]byte("test upload file"))).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m)
				if m.File.FileName == "test2.txt" {
					t.Fatalf("except filename not is test2.txt but got %v", m.File.FileName)
					return
				}
				return
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
