package example_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/qor5/admin/media/oss"
	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/admin/pagebuilder/example"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	publish_view "github.com/qor5/admin/publish/views"
	"github.com/qor5/x/perm"
	"github.com/theplant/gofixtures"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("DBURL")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db.Debug()
}

func TestEditor(t *testing.T) {
	db := ConnectDB()
	b := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin")
	pb := example.ConfigPageBuilder(db, "/page_builder", "", b.I18n())

	sdb, _ := db.DB()
	gofixtures.Data(
		gofixtures.Sql(`
INSERT INTO page_builder_pages ( version, locale_code, title, slug) VALUES ( 'v1','International', '123', '123');
INSERT INTO page_builder_containers ( page_id, page_version,locale_code, model_name, model_id, display_order) VALUES (  1, 'v1','International', 'Header', 1, 1);
INSERT INTO container_headers (color) VALUES ('black');
`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}),
	).TruncatePut(sdb)

	r := httptest.NewRequest("GET", "/page_builder/editors/1?version=v1&locale=International", nil)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	if strings.Index(w.Body.String(), "headers") < 0 {
		t.Error(w.Body.String())
	}

	_, err := pb.AddContainerToPage(1, "v1", "International", "Header")
	if err != nil {
		t.Error(err)
	}

}

func TestUpdatePage(t *testing.T) {
	db := ConnectDB()
	pb := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin").
		Permission(
			perm.New().Policies(
				perm.PolicyFor("root").WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*"),
			).SubjectsFunc(func(r *http.Request) []string {
				return []string{"root"}
			}),
		)
	pageBuilder := example.ConfigPageBuilder(db, "", "", pb.I18n())
	publisher := publish.New(db, oss.Storage).WithPageBuilder(pageBuilder)
	mb := pageBuilder.Configure(pb, db, nil, nil)
	publish_view.Configure(pb, db, nil, publisher, mb)

	// _ = publisher
	pb.Model(&pagebuilder.Page{})

	sdb, _ := db.DB()
	gofixtures.Data(
		gofixtures.Sql(`
INSERT INTO page_builder_pages (id,version, title, slug) VALUES (1,'v1', '123', '123');
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
