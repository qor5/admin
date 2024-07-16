package publish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
)

var dataSeedForFlowDeleteVersion = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES 
('32', '2024-07-04 10:04:04.515824+00', '2024-07-04 10:08:28.045062+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v06', '2024-07-04-x06', '2024-07-04-v05'),
('32', '2024-07-04 10:04:01.725379+00', '2024-07-04 10:04:01.7219+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v05', '2024-07-04-v05', '2024-07-04-v04'),
('32', '2024-07-04 10:03:59.413181+00', '2024-07-04 10:08:35.176457+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v04', '2024-07-04-x04', '2024-07-04-v03'),
('32', '2024-07-04 10:03:57.092015+00', '2024-07-04 10:03:57.088784+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v03', '2024-07-04-v03', '2024-07-04-v02'),
('32', '2024-07-04 10:03:54.402772+00', '2024-07-04 10:03:54.399244+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v02', '2024-07-04-v02', '2024-07-04-v01'),
('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 10:03:29.389412+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v01', '2024-07-04-v01', '');
`, []string{"with_publish_products"}))

type FlowDeleteVersion struct {
	*Flow
}

func TestFlowDeleteVersion(t *testing.T) {
	dataSeedForFlowDeleteVersion.TruncatePut(SQLDB)

	flowDeleteVersion(t, &FlowDeleteVersion{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},
	})
}

func flowDeleteVersion(t *testing.T, f *FlowDeleteVersion) {
	// Add a new resource to test whether the current case will be affected
	flowNew(t, &FlowNew{
		Flow:  f.Flow,
		Name:  "TheTroublemakerProduct",
		Price: 1031,
	})

	selected := "32_2024-07-04-v05"
	id, _ := MustIDVersion(selected)

	toDelete := "32_2024-07-04-v04"
	_, toDeleteVersion := MustIDVersion(toDelete)
	require.NoError(t, f.db.Where("id = ? AND version = ?", id, toDeleteVersion).First(&examples_admin.WithPublishProduct{}).Error)

	models := []*examples_admin.WithPublishProduct{}
	require.NoError(t, f.db.Where("id = ?", id).Order("version DESC").Find(&models).Error)
	require.Len(t, models, 6)

	flowDeleteVersion_Step00_Event_presets_OpenListingDialog(t, f).ThenValidate(
		EnsureVersionListDisplay(selected, models),
	)

	// flowDeleteVersion_Step01_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step02_Event_publish_eventDeleteVersion(t, f).ThenEvent(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, e multipartestutils.TestEventResponse) {
		require.Contains(t, e.RunScript, `emit("PresetsNotifModelsDeletedexamplesAdminWithPublishProduct", {"ids":["`+toDelete+`"]`)
		require.Contains(t, e.RunScript, `"next_version_slug":"32_2024-07-04-v03"`)
	})
	require.ErrorIs(t, f.db.Where("id = ? AND version = ?", id, toDeleteVersion).First(&examples_admin.WithPublishProduct{}).Error, gorm.ErrRecordNotFound)
	require.NoError(t, f.db.Where("id = ?", id).Order("version DESC").Find(&models).Error)
	require.Len(t, models, 5)

	// check resource list draft count
	flowDeleteVersion_Step03_Event___dispatch_stateful_action__(t, f).ThenValidate(
		testflow.ContainsInOrderAtUpdatePortal(0, "<tr", "32_2024-07-04-v06", "FirstProduct", "123", "5", "/tr>"),
	)

	flowDeleteVersion_Step04_Event___dispatch_stateful_action__(t, f).ThenValidate(
		EnsureVersionListDisplay(selected, models),
	)
}

func flowDeleteVersion_Step00_Event_presets_OpenListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_OpenListingDialog").
		Query("f_select_id", "32_2024-07-04-v05").
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
	assert.Equal(t, "presets_listingDialogPortalName", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, testflow.RemoveTime(`setTimeout(function(){ vars.presetsListingDialog = true }, 100)`), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step01_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("id", "32_2024-07-04-v04").
		Query("overlay", "dialog").
		Query("version_name", "2024-07-04-x04").
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
	assert.Equal(t, "deleteConfirm", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step02_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("id", "32_2024-07-04-v04").
		Query("overlay", "dialog").
		Query("version_name", "2024-07-04-x04").
		AddField("VersionName", "2024-07-04-x04").
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
	// assert.Equal(t, testflow.RemoveTime(`locals.deleteConfirmation = false; plaid().vars(vars).emit("PresetsNotifModelsDeletedexamplesAdminWithPublishProduct", {"ids":["32_2024-07-04-v04"]}, {"next_version_slug":"32_2024-07-04-v03"})`), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step03_Event___dispatch_stateful_action__(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("__dispatch_stateful_action__").
		AddField("VersionName", "2024-07-04-x04").
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

func flowDeleteVersion_Step04_Event___dispatch_stateful_action__(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("__dispatch_stateful_action__").
		AddField("__action__", `
{
	"compo_type": "*presets.ListingCompo",
	"compo": {
		"id": "examplespublish_examplewith_publish_products_version_list_dialog_dialog",
		"popup": true,
		"long_style_search_box": true,
		"selected_ids": [],
		"keyword": "",
		"order_bys": null,
		"page": 0,
		"per_page": 0,
		"display_columns": null,
		"active_filter_tab": "",
		"filter_query": "f_select_id=32_2024-07-04-v05",
		"on_mounted": "\n\tvar listingDialogElem = el.ownerDocument.getElementById(\"ListingCompo_examplespublish_examplewith_publish_products_version_list_dialog_dialog\"); \n\tif (listingDialogElem && listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n\t\tlistingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n\t};"
	},
	"injector": "examplespublish_examplewith_publish_products_version_list_dialog",
	"sync_query": false,
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
	assert.Equal(t, "ListingCompo_examplespublish_examplewith_publish_products_version_list_dialog_dialog", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}
