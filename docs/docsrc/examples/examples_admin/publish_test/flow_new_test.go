package publish_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
)

var dataEmptyForFlowNew = gofixtures.Data(gofixtures.Sql(``, []string{"with_publish_products"}))

type FlowNew struct {
	*Flow

	Name  string
	Price int
}

func TestFlowNew(t *testing.T) {
	dataEmptyForFlowNew.TruncatePut(SQLDB)

	flowNew(t, &FlowNew{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},
		Name:  "TestProduct",
		Price: 234,
	})
}

func flowNew(t *testing.T, f *FlowNew) {
	previous := time.Now()

	flowNew_Step00_Event_presets_New(t, f).Then(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
		assert.False(t, ContainsVersionBar(w.Body.String()))
	})

	flowNew_Step01_Event_presets_Update(t, f)

	var m examples_admin.WithPublishProduct
	assert.NoError(t, f.db.Where("created_at > ?", previous).Order("created_at DESC").First(&m).Error)
	assert.Equal(t, m.Version.Version, m.Version.VersionName)
	assert.Empty(t, m.Version.ParentVersion)

	slug := fmt.Sprintf("%d_%s-v01", m.ID, time.Now().Local().Format("2006-01-02"))
	assert.Equal(t, slug, m.PrimarySlug())

	flowNew_Step02_Event___dispatch_stateful_action__(t, f).ThenValidate(
		testflow.ContainsInOrderAtUpdatePortal(0, "<tr", slug, f.Name, fmt.Sprint(f.Price), "1", "/tr>"),
	)

	// other compare
	m.Model = gorm.Model{}
	m.Version = publish.Version{}
	assert.Equal(t, examples_admin.WithPublishProduct{
		Name:     f.Name,
		Price:    f.Price,
		Status:   publish.Status{Status: publish.StatusDraft},
		Schedule: publish.Schedule{},
		Version:  publish.Version{},
	}, m)
}

func flowNew_Step00_Event_presets_New(t *testing.T, f *FlowNew) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_New").
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.PageTitle)
	assert.False(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Len(t, resp.UpdatePortals, 1)
	assert.Equal(t, "presets_RightDrawerPortalName", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, testflow.RemoveTime(`setTimeout(function(){ vars.presetsRightDrawer = true,vars.confirmDrawerLeave=false,vars.presetsDataChanged = {} }, 100)`), testflow.RemoveTime(resp.RunScript))

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("New WithPublishProduct"),
	)

	return testflow.NewThen(t, w, r)
}

func flowNew_Step01_Event_presets_Update(t *testing.T, f *FlowNew) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_Update").
		AddField("Name", f.Name).
		AddField("Price", fmt.Sprint(f.Price)).
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.PageTitle)
	assert.False(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	// assert.Equal(t, testflow.RemoveTime(`plaid().vars(vars).emit("PresetsNotifModelsCreatedexamplesAdminWithPublishProduct", {"models":[{"ID":1,"CreatedAt":"2024-07-04T18:03:29.389412+08:00","UpdatedAt":"2024-07-04T18:03:29.389412+08:00","DeletedAt":null,"Name":"FirstProduct","Price":123,"Status":"draft","OnlineUrl":"","ScheduledStartAt":null,"ScheduledEndAt":null,"ActualStartAt":null,"ActualEndAt":null,"Version":"2024-07-04-v01","VersionName":"2024-07-04-v01","ParentVersion":""}]}); vars.presetsRightDrawer = false; vars.presetsMessage = { show: true, message: "Successfully Updated", color: "success"}`), testflow.RemoveTime(resp.RunScript))
	assert.Contains(t, resp.RunScript, `emit("PresetsNotifModelsCreatedexamplesAdminWithPublishProduct"`)

	return testflow.NewThen(t, w, r)
}

func flowNew_Step02_Event___dispatch_stateful_action__(t *testing.T, f *FlowNew) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("__dispatch_stateful_action__").
		AddField("__action__", `
{
	"compo_type": "*presets.ListingCompo",
	"compo": {
		"id": "examplespublish_examplewith_publish_products_page",
		"popup": false,
		"long_style_search_box": false,
		"selected_ids": [],
		"keyword": "",
		"order_bys": null,
		"page": 0,
		"per_page": 0,
		"display_columns": null,
		"active_filter_tab": "",
		"filter_query": "",
		"on_mounted": ""
	},
	"injector": "examplespublish_examplewith_publish_products",
	"sync_query": true,
	"method": "OnReload",
	"request": {}
}`).
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.PageTitle)
	assert.False(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Len(t, resp.UpdatePortals, 1)
	assert.Equal(t, "ListingCompo_examplespublish_examplewith_publish_products_page", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}
