package example_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor/qor5/media/oss"
	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"

	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	"github.com/qor/qor5/pagebuilder"
	"github.com/qor/qor5/pagebuilder/example"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testingutils"
)

func TestEditor(t *testing.T) {
	db := example.ConnectDB()
	pb := example.ConfigPageBuilder(db, "/page_builder", "")

	sdb, _ := db.DB()
	gofixtures.Data(
		gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, title, slug) VALUES (1, '123', '123');
INSERT INTO public.page_builder_containers (id, page_id, name, model_id, display_order) VALUES (1, 1, 'text_and_image', 1, 0);
INSERT INTO public.page_builder_containers (id, page_id, name, model_id, display_order) VALUES (2, 1, 'text_and_image', 1, 16);
INSERT INTO public.page_builder_containers (id, page_id, name, model_id, display_order) VALUES (3, 1, 'main_content', 1, 40);
INSERT INTO public.text_and_images (text, image, id) VALUES ('Hello Text and Image', null, 1);
`, []string{"page_builder_pages", "page_builder_containers", "text_and_images"}),
	).TruncatePut(sdb)

	for _, oc := range orderCases {
		t.Run(oc.name, func(t *testing.T) {
			err := pb.MoveContainerOrder(1, "", oc.containerID, oc.direction)
			if err != nil {
				t.Error(err)
			}
			var actual []float64
			err = db.Model(&pagebuilder.Container{}).
				Order("id ASC").Pluck("display_order", &actual).Error
			if err != nil {
				t.Error(err)
			}
			if diff := testingutils.PrettyJsonDiff(oc.expected, actual); len(diff) > 0 {
				t.Error(diff)
			}
		})

	}

	r := httptest.NewRequest("GET", "/page_builder/editors/1", nil)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	if strings.Index(w.Body.String(), "main_content") < 0 {
		t.Error(w.Body.String())
	}

	_, err := pb.AddContainerToPage(1, "", "text_and_image")
	if err != nil {
		t.Error(err)
	}

}

var orderCases = []struct {
	name        string
	containerID int
	direction   string
	expected    []float64
}{
	{
		name:        "move 2 up",
		containerID: 2,
		direction:   "up",
		expected:    []float64{0, -8, 40},
	},
	{
		name:        "move 2 up again",
		containerID: 2,
		direction:   "up",
		expected:    []float64{0, -8, 40},
	},
	{
		name:        "move 2 down",
		containerID: 2,
		direction:   "down",
		expected:    []float64{0, 20, 40},
	},
	{
		name:        "move 2 down again",
		containerID: 2,
		direction:   "down",
		expected:    []float64{0, 48, 40},
	},
	{
		name:        "move 2 down twice",
		containerID: 2,
		direction:   "down",
		expected:    []float64{0, 48, 40},
	},
}

func TestUpdatePage(t *testing.T) {
	db := example.ConnectDB()
	pb := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin")
	pageBuilder := example.ConfigPageBuilder(db, "", "")
	publisher := publish.New(db, oss.Storage).WithPageBuilder(pageBuilder)
	mb := pageBuilder.Configure(pb, db)
	publish_view.Configure(pb, db, nil, publisher, mb)

	// _ = publisher
	pb.Model(&pagebuilder.Page{})

	sdb, _ := db.DB()
	gofixtures.Data(
		gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, title, slug) VALUES (1, '123', '123');
`, []string{"page_builder_pages"}),
	).TruncatePut(sdb)

	body := bytes.NewBuffer(nil)

	mw := multipart.NewWriter(body)
	_ = mw.WriteField("__event_data__", `{"eventFuncId":{"id":"presets_Update","params":["1"],"pushState":null},"event":{}}`)
	_ = mw.Close()

	r := httptest.NewRequest("POST", "/admin/pages?__execute_event__=presets_Update", body)
	r.Header.Add("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", mw.Boundary()))

	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)

}
