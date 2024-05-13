package example_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/seo"
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

	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.AddContainerEvent).
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		AddField("containerName", "Header").
		AddField("modelName", "Header").
		AddField("modelID", "1").
		BuildEventFuncRequest()

	bs, _ := httputil.DumpRequest(r, true)
	fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	// var er web.EventResponse
	// _ = json.Unmarshal(w.Body.Bytes(), &er)
	// fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "/page_builder/headers") < 0 {
		t.Error(w.Body.String())
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

func initPageBuilder() (*gorm.DB, *pagebuilder.Builder) {
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
	return db, pb
}

func TestAddContainer(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.AddContainerEvent).
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		AddField("containerName", "Header").
		AddField("modelName", "Header").
		AddField("modelID", "1").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	// var er web.EventResponse
	// _ = json.Unmarshal(w.Body.Bytes(), &er)
	// fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "/page_builder/headers") < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorDeleteContainerConfirmationEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.DeleteContainerConfirmationEvent).
		AddField("containerID", "1_International").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	// fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), presets.DeleteConfirmPortalName) < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorDeleteContainerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.DeleteContainerEvent).
		AddField("containerID", "1_International").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	var er web.EventResponse
	_ = json.Unmarshal(w.Body.Bytes(), &er)
	// fmt.Printf("%#+v\n", er)
	if er.PushState == nil {
		t.Error(w.Body.String())
	}
}

func TestEditorMoveUpDownContainerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.MoveUpDownContainerEvent).
		AddField("containerID", "1_International").
		AddField("moveDirection", "down").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), pagebuilder.ReloadRenderPageOrTemplateEvent) < 0 {
		t.Error(w.Body.String())
	}

	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.MoveUpDownContainerEvent).
		AddField("containerID", "1_International").
		AddField("moveDirection", "up").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)

	if strings.Index(w.Body.String(), pagebuilder.ReloadRenderPageOrTemplateEvent) < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorReloadRenderPageOrTemplateEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.ReloadRenderPageOrTemplateEvent).
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "editorPreviewContentPortal") < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorShowEditContainerDrawerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.ShowEditContainerDrawerEvent).
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		AddField("modelName", "Header").
		AddField("containerName", "Header").
		AddField("modelID", "1").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "pageBuilderRightContentPortal") < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorRenameContainerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.RenameContainerEvent).
		AddField("containerID", "1_International").
		AddField("DisplayName", "Header0000001").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	var er web.EventResponse
	_ = json.Unmarshal(w.Body.Bytes(), &er)
	// fmt.Printf("%#+v\n", er)
	if er.PushState == nil {
		t.Error(w.Body.String())
		return
	}
	r = httptest.NewRequest("GET", "/page_builder/editors/1?pageVersion=v1&locale=International", nil)
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	if strings.Index(w.Body.String(), "Header0000001") < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorShowSortedContainerDrawerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.ShowSortedContainerDrawerEvent).
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		AddField("status", "draft").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "pageBuilderRightContentPortal") < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorMoveContainerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.MoveContainerEvent).
		AddField("moveResult", `[{"container_id":"2","locale":"International"},{"container_id":"1","locale":"International"}]`).
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), pagebuilder.ReloadRenderPageOrTemplateEvent) < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorToggleContainerVisibilityEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.ToggleContainerVisibilityEvent).
		AddField("containerID", "1_International").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), pagebuilder.ReloadRenderPageOrTemplateEvent) < 0 && strings.Index(w.Body.String(), pagebuilder.ShowSortedContainerDrawerEvent) < 0 {
		t.Error(w.Body.String())
	}
}

func TestEditorShowAddContainerDrawerEvent(t *testing.T) {
	var (
		_, pb = initPageBuilder()
		r     *http.Request
		w     *httptest.ResponseRecorder
	)
	r = multipartestutils.NewMultipartBuilder().
		PageURL("/page_builder/editors/1").
		EventFunc(pagebuilder.ShowAddContainerDrawerEvent).
		AddField("containerID", "1_International").
		AddField("pageVersion", "v1").
		AddField("locale", "International").
		BuildEventFuncRequest()

	// bs, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(bs))
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	//var er web.EventResponse
	//_ = json.Unmarshal(w.Body.Bytes(), &er)
	//fmt.Printf("%#+v\n", er)
	if strings.Index(w.Body.String(), "pageBuilderRightContentPortal") < 0 {
		t.Error(w.Body.String())
	}
}
