package integration_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/qor5/x/v3/oss/filesystem"

	"github.com/qor5/admin/v3/media/media_library"
	media_oss "github.com/qor5/admin/v3/media/oss"

	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/media"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/role"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
)

var mediaTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, folder,file) VALUES (1, '2024-06-14 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', false,'{"FileName":"Snipaste_2024-06-14_10-06-12.png","Url":"/system/media_libraries/1/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
INSERT INTO public.media_libraries (id,user_id, created_at, updated_at, deleted_at, selected_type, folder,file) VALUES (2, 888, '2024-06-15 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', false, '{"FileName":"test_search1.png","Url":"/system/media_libraries/2/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
INSERT INTO public.media_libraries (id,user_id, created_at, updated_at, deleted_at, selected_type,folder, file) VALUES (3, 999,'2024-06-14 02:06:15.811153 +00:00', '2024-06-19 08:25:12.954584 +00:00', null, 'image', false, '{"FileName":"test_search2.png","Url":"/system/media_libraries/3/file.png","Width":1598,"Height":966,"FileSizes":{"@qor_preview":7140,"default":128870,"original":128870},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0},"og":{"Width":1200,"Height":630,"Padding":false,"Sm":0,"Cols":0},"twitter-large":{"Width":1200,"Height":600,"Padding":false,"Sm":0,"Cols":0},"twitter-small":{"Width":630,"Height":630,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"123123"}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (4, '2024-07-26 02:17:18.957978 +00:00', '2024-07-26 02:17:18.957978 +00:00', null, '', '{"FileName":"test001","Url":"","Video":"","SelectedType":"","Description":""}', 888, true, 0);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (5, '2024-07-26 02:17:18.957978 +00:00', '2024-07-26 02:17:18.957978 +00:00', null, '', '{"FileName":"test001","Url":"","Video":"","SelectedType":"","Description":""}', 888, true, 4);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (6, '2024-07-26 02:17:18.957978 +00:00', '2024-07-26 02:17:18.957978 +00:00', null, '', '{"FileName":"test.png","Url":"","Video":"","SelectedType":"","Description":""}', 888, true, 0);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (7, '2024-07-26 02:17:18.957978 +00:00', '2024-07-26 02:17:18.957978 +00:00', null, 'video', '{"FileName":"test.mp4","Url":"","Video":"","SelectedType":"","Description":""}', 888, false, 0);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (8, '2024-07-26 02:17:18.957978 +00:00', '2024-07-26 02:17:18.957978 +00:00', null, 'file', '{"FileName":"test.txt","Url":"","Video":"","SelectedType":"","Description":""}', 888, false, 0);

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
					PageURL("/pages/1_2024-06-19-v01_Japan?__execute_event__=mediaLibrary_ChooseFileEvent&field=SEO.OpenGraphImageFromMediaLibrary&media_ids=1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Snipaste_2024-06-14_10-06-12.png"},
		},
		{
			Name:  "LoadImageCropperEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/1_2024-06-19-v01_Japan").
					EventFunc("mediaLibrary_LoadImageCropperEvent").
					AddField("media_ids", "1").
					AddField("thumb", "default").
					AddField("field", "Image").
					AddField("cfg", `{"Sizes":null,"Max":0,"AllowType":"","BackgroundColor":"","DisableCrop":false,"SimpleIMGURL":false}`).
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{":view-mode='2'"},
		},
		{
			Name:  "MediaLibrary Admin Role List",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					AddField("type", "all").
					AddField("order_by", "created_at_desc").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test_search1.png", "test_search2.png"},
		},
		{
			Name:  "MediaLibrary Admin Role List asc",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					AddField("type", "all").
					AddField("order_by", "created_at").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test_search2.png", "test_search1.png"},
		},
		{
			Name:  "MediaLibrary Create Folder",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.CreateFolderEvent).
					AddField(media.ParamName, "test_create_directory").
					AddField(media.ParamParentID, "0").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m *media_library.MediaLibrary
				if err := TestDB.Order("id desc").First(&m).Error; err != nil {
					t.Fatalf("create directory err : %v", err)
				}
				if !m.Folder || m.File.FileName != "test_create_directory" {
					t.Fatalf("create directory : %#+v", m)
				}
			},
		},
		{
			Name:  "MediaLibrary Create Folder Empty ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.CreateFolderEvent).
					AddField(media.ParamName, "").
					AddField(media.ParamParentID, "0").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"folder name can`t be empty"},
		},
		{
			Name:  "MediaLibrary New Folder Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.NewFolderDialogEvent).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"vx-dialog", "New Folder"},
		},
		{
			Name:  "MediaLibrary Move To Folder Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.MoveToFolderDialogEvent).
					Query(media.ParamSelectIDS, "1,2,3").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"vx-dialog", "Choose Folder", "Root Directory", "0_folder_portal_name"},
			ExpectPortalUpdate0NotContains:     []string{"test001"},
		},
		{
			Name:  "MediaLibrary Next Folder",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.NextFolderEvent).
					Query(media.ParamSelectFolderID, "0").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test001"},
			ExpectPortalUpdate0NotContains:     []string{"test002"},
		},

		{
			Name:  "MediaLibrary Move To Folder",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.MoveToFolderEvent).
					Query(media.ParamSelectFolderID, "5").
					Query(media.ParamSelectIDS, "1,2,3").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				if err := TestDB.Where("id in (1,2,3) and parent_id=5 ").Model(media_library.MediaLibrary{}).Count(&count).Error; err != nil {
					t.Fatalf("move to folder err : %v", err)
				}
				if count != 3 {
					t.Fatalf("move to folder count : %d", count)
				}
			},
		},
		{
			Name:  "MediaLibrary Delete Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.DeleteConfirmationEvent).
					Query(media.ParamMediaIDS, "1,2,3").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"vx-dialog", "Are you sure you want to delete"},
		},
		{
			Name:  "MediaLibrary Delete One object",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.DoDeleteEvent).
					Query(media.ParamMediaIDS, "1").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				if err := TestDB.Model(media_library.MediaLibrary{}).Where("id=1").Count(&count).Error; err != nil {
					t.Fatalf("delete object err : %v", err)
				}
				if count != 0 {
					t.Fatalf("not delete object count : %d", count)
				}
			},
		},
		{
			Name:  "MediaLibrary Delete parent folder",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.DoDeleteEvent).
					Query(media.ParamMediaIDS, "4").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				if err := TestDB.First(&m, 5).Error; err != nil {
					t.Fatalf("find object err : %v", err)
				}
				if m.ParentId != 0 {
					t.Fatalf("not update no parent folder %v", m.ParentId)
				}
			},
		},
		{
			Name:  "MediaLibrary Delete Objects",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.DoDeleteEvent).
					Query(media.ParamMediaIDS, "1,2,3,4,5,6").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				if err := TestDB.Model(media_library.MediaLibrary{}).Where("id in (1,2,3,4,5,6)").Count(&count).Error; err != nil {
					t.Fatalf("delete objects err : %v", err)
				}
				if count != 0 {
					t.Fatalf("not delete objects count : %d", count)
				}
			},
		},
		{
			Name:  "MediaLibrary Rename Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.RenameDialogEvent).
					Query(media.ParamMediaIDS, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Rename", "Name", "Snipaste_2024-06-14_10-06-12"},
			ExpectPortalUpdate0NotContains:     []string{".png"},
		},
		{
			Name:  "MediaLibrary Rename Folder Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.RenameDialogEvent).
					Query(media.ParamMediaIDS, "6").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Rename", "Name", "test.png"},
		},
		{
			Name:  "MediaLibrary Rename",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.RenameEvent).
					Query(media.ParamMediaIDS, "1").
					AddField(media.ParamName, "1").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				if err := TestDB.First(&m, "1").Error; err != nil {
					t.Fatalf("rename err : %v", err)
				}
				if m.File.FileName != "1.png" {
					t.Fatalf("rename failed need:<1.png>,got <%s>", m.File.FileName)
				}
			},
		},
		{
			Name:  "MediaLibrary Rename Folder",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.RenameEvent).
					Query(media.ParamMediaIDS, "6").
					AddField(media.ParamName, "test").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				if err := TestDB.First(&m, "6").Error; err != nil {
					t.Fatalf("rename err : %v", err)
				}
				if m.File.FileName != "test" {
					t.Fatalf("rename failed need:<test>,got <%s>", m.File.FileName)
				}
			},
		},
		{
			Name:  "MediaLibrary Update Description Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.UpdateDescriptionDialogEvent).
					Query(media.ParamMediaIDS, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Update Description", "Description", "123123"},
		},
		{
			Name:  "MediaLibrary Update Description",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.UpdateDescriptionEvent).
					Query(media.ParamMediaIDS, "1").
					AddField(media.ParamCurrentDescription, "321").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				if err := TestDB.First(&m, "1").Error; err != nil {
					t.Fatalf("update description err : %v", err)
				}
				if m.File.Description != "321" {
					t.Fatalf("update description failed need:<321>,got <%s>", m.File.Description)
				}
			},
		},
		{
			Name:  "MediaLibrary Open FileChooser",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"v-dialog", "Choose File"},
		},
		{
			Name:  "MediaLibrary Search",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query(web.EventFuncIDName, media.ImageSearchEvent).
					Query("keyword", "test_search2").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test_search2.png"},
			ExpectPortalUpdate0NotContains:     []string{"test_search1.png"},
		},
		{
			Name:  "MediaLibrary Search file name",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query(web.EventFuncIDName, media.ImageSearchEvent).
					Query("keyword", "2").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test_search2.png"},
			ExpectPortalUpdate0NotContains:     []string{"test_search1.png"},
		},
		{
			Name:  "MediaLibrary Jump",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query("page", "2").
					Query(web.EventFuncIDName, media.ImageJumpPageEvent).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0NotContains: []string{"test_search2.png", "test_search1.png"},
		},
		{
			Name:  "MediaLibrary Folder Tab Select Image Type ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("type", "image").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test001", "test_search1.png", "test_search2.png"},
			ExpectPageBodyNotContains:     []string{"test.mp4", "test.txt"},
		},
		{
			Name:  "Pages Folder Tab Cfg Just allow image",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("cfg", `{"AllowType":"image"}`).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test001", "test_search1.png", "test_search2.png"},
			ExpectPortalUpdate0NotContains:     []string{"test.mp4", "test.txt"},
		},
		{
			Name:  "MediaLibrary Folder Tab Select Video Type",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("type", "video").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test001", "test.mp4"},
			ExpectPageBodyNotContains:     []string{"test_search1.png", "test_search2.png", "test.txt"},
		},
		{
			Name:  "Pages Folder Tab Cfg Just allow video",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("cfg", `{"AllowType":"video"}`).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test001", "test.mp4"},
			ExpectPortalUpdate0NotContains:     []string{"test_search1.png", "test_search2.png", "test.txt"},
		},
		{
			Name:  "MediaLibrary Folder Tab Select file Type ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("type", "file").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"test001", "test.txt"},
			ExpectPageBodyNotContains:     []string{"test.mp4", "test_search1.png", "test_search2.png"},
		},
		{
			Name:  "Pages Folder Tab Cfg Just allow file",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					Query(media.ParamField, "media").
					Query("tab", "folders").
					Query("cfg", `{"AllowType":"file"}`).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test001", "test.txt"},
			ExpectPortalUpdate0NotContains:     []string{"test.mp4", "test_search1.png", "test_search2.png"},
		},
		{
			Name:  "MediaLibrary folders Tab upload file",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
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
				media_oss.Storage = filesystem.New("/tmp/media_test")
				if m.File.FileName != "test2.txt" {
					t.Fatalf("except filename: test2.txt but got %v", m.File.FileName)
				}
				if m.UserID != 888 {
					t.Fatalf("except user_id: 888 but got %v", m.UserID)
				}
			},
		},
		{
			Name:  "Pages Folder Tab upload file",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				media_oss.Storage = filesystem.New("/tmp/media_test")
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.UploadFileEvent).
					Query(media.ParamField, "media").
					AddReader("NewFiles", "test2.txt", bytes.NewReader([]byte("test upload file"))).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m media_library.MediaLibrary
				TestDB.Order("id desc").First(&m)
				if m.File.FileName != "test2.txt" {
					t.Fatalf("except filename: test2.txt but got %v", m.File.FileName)
				}
			},
		},
		{
			Name:  "Pages ChooseFileEvent Dialog no selected image",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					Query(media.ParamField, "media").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"select_ids:[]", "v-checkbox"},
			ExpectPageBodyNotContains:          []string{"Save"},
		},

		{
			Name:  "MediaLibrary Folder Tab Has Parent Fold",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				mediaTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/media-libraries").
					Query(web.EventFuncIDName, media.OpenFileChooserEvent).
					Query(media.ParamField, "media").
					Query(media.ParamParentID, "4").
					Query("tab", "folders").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"v-breadcrumbs"},
			ExpectPortalUpdate0NotContains:     []string{"test.mp4", "test_search1.png", "test_search2.png"},
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
					PageURL("/media-libraries").
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
