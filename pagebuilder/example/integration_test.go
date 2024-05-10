package example_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/theplant/gofixtures"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/login"

	"github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/perm"
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
INSERT INTO page_builder_pages (id, version, locale_code, title, slug) VALUES (1, 'v1','International', '123', '123');
INSERT INTO page_builder_containers ( page_id, page_version,locale_code, model_name, model_id, display_order) VALUES (  1, 'v1','International', 'Header', 1, 1);
INSERT INTO container_headers (color) VALUES ('black');
`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}),
	).TruncatePut(sdb)

	r := httptest.NewRequest("GET", "/page_builder/editors/1?pageVersion=v1&locale=International", nil)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	if strings.Index(w.Body.String(), "headers") < 0 {
		t.Error(w.Body.String())
	}

	_, err := pb.AddContainerToPage(1, "", "v1", "International", "Header")
	if err != nil {
		t.Error(err)
	}

}

func TestUpdatePage(t *testing.T) {
	db := ConnectDB()
	ab := activity.New(db).CreatorContextKey(login.UserKey).TabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		})
	pb := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin").
		Permission(
			perm.New().Policies(
				perm.PolicyFor("root").WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*"),
			).SubjectsFunc(func(r *http.Request) []string {
				return []string{"root"}
			}),
		)
	pageBuilder := example.ConfigPageBuilder(db, "", "", pb.I18n())
	publisher := publish.New(db, oss.Storage)
	pageBuilder.SEO(seo.New(db)).Publisher(publisher).Activity(ab)
	pageBuilder.Install(pb)
	publisher.Install(pb)

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

func initPageBuilder() (*gorm.DB, *pagebuilder.Builder, *web.EventContext) {
	db, err := gorm.Open(postgres.Open(os.Getenv("DBURL")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	b := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin")
	pb := example.ConfigPageBuilder(db, "/page_builder", "", b.I18n())
	sdb, _ := db.DB()
	p := pagebuilder.Page{}
	p.L10nON()
	gofixtures.Data(
		gofixtures.Sql(`
INSERT INTO page_builder_pages (id, version, locale_code, title, slug) VALUES (1, 'v1','International', '123', '123');
INSERT INTO page_builder_containers (id, page_id, page_version,locale_code, model_name, model_id, display_order) VALUES ( 1, 1, 'v1','International', 'Header', 1, 1),( 2, 1, 'v1','International', 'Header', 1, 2);
INSERT INTO container_headers (id,color) VALUES (1,'black'),(2,'black');
`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}),
	).TruncatePut(sdb)
	ctx := &web.EventContext{
		R: &http.Request{
			URL:  new(url.URL),
			Form: url.Values{},
		},
	}
	return db, pb, ctx
}

func TestAddContainer(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("id", "1")
	ctx.R.Form.Set("pageVersion", "v1")
	ctx.R.Form.Set("locale", "International")
	ctx.R.Form.Set("modelName", "Header")
	ctx.R.Form.Set("containerName", "Header")
	ctx.R.Form.Set("modelID", "1")
	if _, err = pb.AddContainer(ctx); err != nil {
		t.Error(err)
		return
	}
}

func TestEditorDelete(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)

	ctx.R.Form.Set("containerID", "1_International")
	if _, err = pb.DeleteContainer(ctx); err != nil {
		t.Error(err)
		return
	}
}

func TestEditorMoveUpDown(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	TestAddContainer(t)
	ctx.R.Form.Set("containerID", "1_International")
	ctx.R.Form.Set("moveDirection", "down")
	if _, err = pb.MoveUpDownContainer(ctx); err != nil {
		t.Error(err)
		return
	}
	ctx.R.Form.Set("moveDirection", "up")
	if _, err = pb.MoveUpDownContainer(ctx); err != nil {
		t.Error(err)
	}
}

func TestReloadRenderPage(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("id", "1")
	ctx.R.Form.Set("pageVersion", "v1")
	ctx.R.Form.Set("locale", "International")
	if _, err = pb.ReloadRenderPageOrTemplate(ctx); err != nil {
		t.Error(err)
	}
}

func TestShowEditContainerDrawer(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("id", "1")
	ctx.R.Form.Set("pageVersion", "v1")
	ctx.R.Form.Set("locale", "International")
	ctx.R.Form.Set("modelName", "Header")
	ctx.R.Form.Set("containerName", "Header")
	ctx.R.Form.Set("modelID", "1")
	if _, err = pb.ShowEditContainerDrawer(ctx); err != nil {
		t.Error(err)
	}
}

func TestRenameContainer(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("containerID", "1_International")
	ctx.R.Form.Set("DisplayName", "Header001")
	if _, err = pb.RenameContainer(ctx); err != nil {
		t.Error(err)
	}
}

func TestShowSortedContainerDrawer(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("id", "1")
	ctx.R.Form.Set("pageVersion", "v1")
	ctx.R.Form.Set("locale", "International")
	ctx.R.Form.Set("status", "draft")
	if _, err = pb.ShowSortedContainerDrawer(ctx); err != nil {
		t.Error(err)
	}
}

func TestMoveContainerEvent(t *testing.T) {
	var (
		err        error
		_, pb, ctx = initPageBuilder()
	)
	ctx.R.Form.Set("moveResult", `[{"container_id":"2","locale":"International"},{"container_id":"1","locale":"International"}]`)
	if _, err = pb.MoveContainer(ctx); err != nil {
		t.Error(err)
	}
}
