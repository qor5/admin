package example_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	m.Run()
}

var pageBuilderData = gofixtures.Data(
	gofixtures.Sql(`
INSERT INTO page_builder_pages (id, version, locale_code, title, slug) VALUES (1, 'v1','International', '123', '123');
INSERT INTO container_headers (id, color) VALUES (1, 'black');
INSERT INTO page_builder_containers (id, page_id, page_model_name, page_version, locale_code, model_name, model_id, 
display_order) VALUES (1, 1, 'pages', 'v1','International', 'Header', 1, 1),(2, 1, 'pages', 'v1','International', 
'Header', 1,
2);
`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}),
)

func initPageBuilder() (*gorm.DB, *pagebuilder.Builder, *presets.Builder) {
	db := TestDB
	b := presets.New().DataOperator(gorm2op.DataOperator(db)).URIPrefix("/admin")
	pb := example.ConfigPageBuilder(db, "/page_builder", "", b.I18n())
	ab := activity.New(db).CreatorContextKey(login.UserKey).TabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		})
	publisher := publish.New(db, oss.Storage)
	pb.Publisher(publisher).SEO(seo.New(db, seo.WithLocales("International"))).Activity(ab)
	b.Use(pb)

	return db, pb, b
}

func TestPages(t *testing.T) {
	db, _, p := initPageBuilder()
	dbr, _ := db.DB()
	cases := []multipartestutils.TestCase{
		{
			Name:  "Update page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/admin/pages").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1_v1_International").
					AddField("Title", "Hello Page").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"success"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, p)
		})
	}
}

func TestPageBuilder(t *testing.T) {
	db, pb, _ := initPageBuilder()
	dbr, _ := db.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Show Editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/editors/1_v1_International", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Header"},
		},
		{
			Name:  "Add Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.AddContainerEvent).
					AddField("containerName", "Header").
					AddField("modelName", "Header").
					AddField("id", "1").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"page_builder_ReloadRenderPageOrTemplateEvent", "pageBuilderRightContentPortal", "overlay", "content"},
		},
		{
			Name:  "Delete Container Confirmation Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.DeleteContainerConfirmationEvent).
					AddField("containerID", "1_International").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{presets.DeleteConfirmPortalName},
		},
		{
			Name:  "Editor Delete Container Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.DeleteContainerEvent).
					AddField("containerID", "1_International").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"pushState(true)", "clearMergeQuery"},
		},
		{
			Name:  "Editor Move Down Container Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					AddField("containerID", "1_International").
					AddField("moveDirection", "down").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{pagebuilder.ReloadRenderPageOrTemplateEvent},
		},
		{
			Name:  "Editor Move Up Container Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					AddField("containerID", "1_International").
					AddField("moveDirection", "up").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{pagebuilder.ReloadRenderPageOrTemplateEvent},
		},
		{
			Name:  "Editor Reload Render Page Or Template Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.ReloadRenderPageOrTemplateEvent).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"vx-scroll-iframe"},
		},
		{
			Name:  "Editor Rename Container Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.RenameContainerEvent).
					AddField("containerID", "1_International").
					AddField("DisplayName", "Header0000001").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{pagebuilder.ShowSortedContainerDrawerEvent, pagebuilder.ReloadRenderPageOrTemplateEvent},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var pc pagebuilder.Container
				db.Find(&pc, 1)
				if pc.DisplayName != "Header0000001" {
					t.Error("Expected Header0000001 got ", pc.DisplayName)
				}
			},
		},
		{
			Name:  "Editor Show Sorted Container Drawer Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.ShowSortedContainerDrawerEvent).
					AddField("status", "draft").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`"Header"`},
		},
		{
			Name:  "Editor Move Container Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.MoveContainerEvent).
					AddField("moveResult", `[{"container_id":"2","locale":"International"},{"container_id":"1","locale":"International"}]`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"/page_builder/pages/editors/1_v1_International"},
		},
		{
			Name:  "Editor Toggle Container Visibility Event",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/page_builder/pages/editors/1_v1_International").
					EventFunc(pagebuilder.ToggleContainerVisibilityEvent).
					AddField("containerID", "1_International").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{
				pagebuilder.ReloadRenderPageOrTemplateEvent,
				pagebuilder.ShowSortedContainerDrawerEvent,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
