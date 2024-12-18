package examples_admin

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	media_oss "github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var simpleMediaData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.input_demos (id, text_field1, text_area1, switch1, slider1, select1, range_slider1, radio1, file_input1, combobox1, checkbox1, autocomplete1, button_group1, chip_group1, item_group1, list_item_group1, slide_group1, color_picker1, date_picker1, date_picker_month1, time_picker1, media_library1, updated_at, created_at) VALUES (1, '', '', false, 0, '', null, '', '', '', false, null, '', '', '', '', '', '', '', '', '', '{"ID":6,"Url":"/system/media_libraries/6/file.png","VideoLink":"","FileName":"截屏2024-07-31 11.18.05.png","Description":"","FileSizes":{"@qor_preview":14413,"default":1037427,"original":1037427},"Width":2868,"Height":2090}', '2024-07-31 09:31:01.955148 +00:00', '2024-07-31 08:09:36.838435 +00:00');

INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (5, '2024-07-31 02:28:11.318737 +00:00', '2024-07-31 02:28:12.874616 +00:00', null, 'image', '{"FileName":"截屏2024-07-31 10.28.05.png","Url":"/system/media_libraries/5/file.png","Width":3230,"Height":1908,"FileSizes":{"@qor_preview":11765,"default":1083448,"original":1083448},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"5_desc"}', 0, false, 0);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (6, '2024-07-31 03:18:11.485960 +00:00', '2024-07-31 03:18:12.527071 +00:00', null, 'image', '{"FileName":"截屏2024-07-31 11.18.05.png","Url":"/system/media_libraries/6/file.png","Width":2868,"Height":2090,"FileSizes":{"@qor_preview":14413,"default":1037427,"original":1037427},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"6_desc"}', 0, false, 0);
`, []string{"media_libraries", "input_demos", "media_roles"}))

func TestMediaExample(t *testing.T) {
	dbr, _ := TestDB.DB()
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	MediaExample(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "default choose media",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/default?__execute_event__=mediaLibrary_ChooseFileEvent").
					Query("cfg", `{"Sizes":null,"Max":0,"AllowType":"","BackgroundColor":"","DisableCrop":false,"SimpleIMGURL":false}`).
					Query("field", "MediaLibrary1").
					Query("media_ids", "5").
					AddField("type", "all").
					AddField("order_by", "created_at_desc").
					BuildEventFuncRequest()
				return req
			},
			// have img information and description
			ExpectPortalUpdate0ContainsInOrder: []string{"/system/media_libraries/5/file.png", "<span>default</span>", "3230 X 1908", `<vx-field v-model='form["MediaLibrary1.Description"]' :error-messages='dash.errorMessages["MediaLibrary1.Description"]' v-assign:append='[dash.errorMessages, {"MediaLibrary1.Description":null}]' v-assign='[form, {"MediaLibrary1.Description":"5_desc"}]' placeholder='description for accessibility' :disabled='false'></vx-field>`},
			ExpectPortalUpdate0NotContains:     []string{"/system/media_libraries/6/file.png"},
		},
		{
			Name:  "simple choose media",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/simple?__execute_event__=mediaLibrary_ChooseFileEvent").
					Query("cfg", `{"Sizes":null,"Max":0,"AllowType":"","BackgroundColor":"","DisableCrop":false,"SimpleIMGURL":true}`).
					Query("field", "MediaLibrary1").
					Query("media_ids", "5").
					AddField("type", "all").
					AddField("order_by", "created_at_desc").
					BuildEventFuncRequest()
				return req
			},
			// not have img information and description
			ExpectPortalUpdate0ContainsInOrder: []string{"/system/media_libraries/5/file.png"},
			ExpectPortalUpdate0NotContains:     []string{"/system/media_libraries/6/file.png", "<span>default</span>", "3230 X 1908"},
		},
		{
			Name:  "default update",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/default?__execute_event__=presets_Update&id=1").
					AddField("MediaLibrary1.Values", "{\"ID\":5,\"Url\":\"/system/media_libraries/5/file.png\",\"VideoLink\":\"\",\"FileName\":\"simple.png\",\"Description\":\"5_desc\",\"FileSizes\":{\"@qor_preview\":11765,\"default\":1083448,\"original\":1083448},\"Width\":3230,\"Height\":1908}").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"/system/media_libraries/5/file.png", "Successfully Updated"},
		},
		{
			Name:  "simple update",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/simple?__execute_event__=presets_Update&id=1").
					AddField("MediaLibrary1.Values", "{\"ID\":5,\"Url\":\"/system/media_libraries/5/file.png\",\"VideoLink\":\"\",\"FileName\":\"simple.png\",\"Description\":\"5_desc\",\"FileSizes\":{\"@qor_preview\":11765,\"default\":1083448,\"original\":1083448},\"Width\":3230,\"Height\":1908}").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"/system/media_libraries/5/file.png", "Successfully Updated"},
		},
		{
			Name:  "upload file wrap saver",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				media_oss.Storage = filesystem.New("/tmp/media_test")
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.UploadFileEvent).
					Query(media.ParamField, "media").
					AddReader("NewFiles", "test2.txt", bytes.NewReader([]byte("test upload file"))).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m)
				if m.File.FileName != "test2.txt" {
					t.Fatalf("except filename: test2.txt but got %v", m.File.FileName)
					return
				}
				var mr MediaRole
				TestDB.Order("id desc").First(&mr)
				if mr.RoleName != "viewer" {
					t.Fatalf("except rolename: viewer but got %v", mr.RoleName)
				}
				return
			},
		},
		{
			Name:  "crop file wrap saver",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, "mediaLibrary_CropImageEvent").
					Query(media.ParamField, "media").
					Query(media.ParamMediaIDS, "5").
					Query("CropOption", ` {"x":70.40625,"y":23.05859375,"width":267.62109375,"height":323.7265625,"rotate":0,"scaleX":1,"scaleY":1}`).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m, 5)
				except := "截屏2024-07-31 10.28.05.png"
				if m.File.FileName != except {
					t.Fatalf("except filename: %s but got %v", except, m.File.FileName)
					return
				}
				var mr MediaRole
				TestDB.Order("id desc").First(&mr)
				if mr.RoleName != "" {
					t.Fatalf("except no rolename  but got %v", mr.RoleName)
				}
				return
			},
		},
		{
			Name:  "create folder wrap saver",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.CreateFolderEvent).
					Query(media.ParamName, "folder_test").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m)
				if m.File.FileName != "folder_test" {
					t.Fatalf("except filename: folder_test but got %v", m.File.FileName)
					return
				}
				var mr MediaRole
				TestDB.Order("id desc").First(&mr)
				if mr.RoleName != "viewer_folder" {
					t.Fatalf("except rolename: viewer_folder but got %v", mr.RoleName)
				}
				return
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
