package examples_admin

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"net/http"
	"testing"
)

var simpleMediaDate = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.input_demos (id, text_field1, text_area1, switch1, slider1, select1, range_slider1, radio1, file_input1, combobox1, checkbox1, autocomplete1, button_group1, chip_group1, item_group1, list_item_group1, slide_group1, color_picker1, date_picker1, date_picker_month1, time_picker1, media_library1, updated_at, created_at) VALUES (1, '', '', false, 0, '', null, '', '', '', false, null, '', '', '', '', '', '', '', '', '', '{"ID":6,"Url":"/system/media_libraries/6/file.png","VideoLink":"","FileName":"截屏2024-07-31 11.18.05.png","Description":"","FileSizes":{"@qor_preview":14413,"default":1037427,"original":1037427},"Width":2868,"Height":2090}', '2024-07-31 09:31:01.955148 +00:00', '2024-07-31 08:09:36.838435 +00:00');

INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (5, '2024-07-31 02:28:11.318737 +00:00', '2024-07-31 02:28:12.874616 +00:00', null, 'image', '{"FileName":"截屏2024-07-31 10.28.05.png","Url":"/system/media_libraries/5/file.png","Width":3230,"Height":1908,"FileSizes":{"@qor_preview":11765,"default":1083448,"original":1083448},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"5_desc"}', 0, false, 0);
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file, user_id, folder, parent_id) VALUES (6, '2024-07-31 03:18:11.485960 +00:00', '2024-07-31 03:18:12.527071 +00:00', null, 'image', '{"FileName":"截屏2024-07-31 11.18.05.png","Url":"/system/media_libraries/6/file.png","Width":2868,"Height":2090,"FileSizes":{"@qor_preview":14413,"default":1037427,"original":1037427},"Sizes":{"default":{"Width":0,"Height":0,"Padding":false,"Sm":0,"Cols":0}},"Video":"","SelectedType":"","Description":"6_desc"}', 0, false, 0);
`, []string{"media_libraries", "input_demos"}))

func TestMediaExample(t *testing.T) {
	dbr, _ := TestDB.DB()
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	MediaExample(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "default choose media",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaDate.TruncatePut(dbr)
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
			ExpectPortalUpdate0ContainsInOrder: []string{"/system/media_libraries/5/file.png", "<span>default</span>", "3230 X 1908", "<v-text-field v-model='form[\"MediaLibrary1.Description\"]' v-assign='[form, {\"MediaLibrary1.Description\":\"5_desc\"}]'"},
			ExpectPortalUpdate0NotContains:     []string{"/system/media_libraries/6/file.png"},
		},
		{
			Name:  "simple choose media",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaDate.TruncatePut(dbr)
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
			ExpectPortalUpdate0NotContains:     []string{"/system/media_libraries/6/file.png", "<span>default</span>", "3230 X 1908", "<v-text-field v-model='form[\"MediaLibrary1.Description\"]' v-assign='[form, {\"MediaLibrary1.Description\":\"5_desc\"}]'"},
		},
		{
			Name:  "default update",
			Debug: true,
			ReqFunc: func() *http.Request {
				simpleMediaDate.TruncatePut(dbr)
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
				simpleMediaDate.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/simple?__execute_event__=presets_Update&id=1").
					AddField("MediaLibrary1.Values", "{\"ID\":5,\"Url\":\"/system/media_libraries/5/file.png\",\"VideoLink\":\"\",\"FileName\":\"simple.png\",\"Description\":\"5_desc\",\"FileSizes\":{\"@qor_preview\":11765,\"default\":1083448,\"original\":1083448},\"Width\":3230,\"Height\":1908}").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"/system/media_libraries/5/file.png", "Successfully Updated"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
