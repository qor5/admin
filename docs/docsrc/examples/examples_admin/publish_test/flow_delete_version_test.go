package publish_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/docs/v3/docsrc/examples/examples_admin"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
)

var dataSeedForFlowDeleteVersion = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('1', '2024-05-31 03:07:57.522522+00', '2024-05-31 03:07:57.519697+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-31-v01', '2024-05-31-v01', '2024-05-26-v06'),
('1', '2024-05-26 13:13:14.463547+00', '2024-05-26 13:14:47.64948+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v04', '2024-05-26-x04', '2024-05-26-v03'),
('1', '2024-05-26 13:13:18.256404+00', '2024-05-26 13:14:43.729016+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v06', '2024-05-26-x06', '2024-05-26-v05'),
('1', '2024-05-26 13:13:16.56434+00', '2024-05-26 13:14:39.705527+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v05', '2024-05-26-x05', '2024-05-26-v04'),
('1', '2024-05-26 13:13:11.858454+00', '2024-05-26 13:13:11.855648+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v03', '2024-05-26-v03', '2024-05-26-v02'),
('1', '2024-05-26 13:13:09.768116+00', '2024-05-26 13:13:09.764082+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v02', '2024-05-26-v02', '2024-05-26-v01'),
('1', '2024-05-26 13:12:06.408234+00', '2024-05-26 13:12:06.408234+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v01', '2024-05-26-v01', '');
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

	displayID := "1_2024-05-31-v01"
	id, _ := MustIDVersion(displayID)

	models := []*examples_admin.WithPublishProduct{}
	updateExpectedModels := func(expectedCount int) {
		require.NoError(t, f.db.Where("id = ?", id).Order("version DESC").Find(&models).Error)
		assert.Len(t, models, expectedCount)
	}
	updateExpectedModels(7)

	selectID := displayID

	ensureListDisplay := func() testflow.ValidatorFunc {
		return EnsureVersionListDisplay(selectID, models)
	}

	// Test: Delete versions that are neither currently selected nor currently displayed
	flowDeleteVersion_Step00_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step01_Event_presets_OpenListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step02_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step03_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(6)

	flowDeleteVersion_Step04_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	// Test: Select another item and delete it to test whether deleting the current selection will reselect the currently displayed item.
	selectID = "1_2024-05-26-v04"
	flowDeleteVersion_Step05_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step06_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step07_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(5)

	selectID = displayID
	flowDeleteVersion_Step08_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	// Test: Select another item but delete the currently displayed one to test whether it will switch the display without changing the selection.
	selectID = "1_2024-05-26-v03"
	flowDeleteVersion_Step09_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step10_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step11_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(4)

	displayID = "1_2024-05-26-v05"
	flowDeleteVersion_Step12_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step13_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	// Test: Confirm the current selection, then reopen and delete it to test whether it will switch both the display and the selection to an older version.
	flowDeleteVersion_Step14_Event_publish_eventSelectVersion(t, f)
	displayID = selectID
	flowDeleteVersion_Step15_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step16_Event_presets_OpenListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step17_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step18_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(3)

	displayID = "1_2024-05-26-v02"
	selectID = displayID
	flowDeleteVersion_Step19_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step20_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	// Test: Select the oldest version and confirm display, then delete it to test whether it will switch to and display the newest version when there are no older versions to switch to.
	selectID = "1_2024-05-26-v01"
	flowDeleteVersion_Step21_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step22_Event_publish_eventSelectVersion(t, f)
	displayID = selectID
	flowDeleteVersion_Step23_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step24_Event_presets_OpenListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step25_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step26_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(2)

	displayID = "1_2024-05-26-v05"
	selectID = displayID
	flowDeleteVersion_Step27_Event_presets_DetailingDrawer(t, f).ThenValidate(EnsureCurrentDisplayID(displayID))

	flowDeleteVersion_Step28_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	// Delete all remaining versions
	flowDeleteVersion_Step29_Event_publish_eventDeleteVersionDialog(t, f)

	flowDeleteVersion_Step30_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(1)

	flowDeleteVersion_Step31_Event_presets_UpdateListingDialog(t, f).ThenValidate(ensureListDisplay())

	flowDeleteVersion_Step32_Event_publish_eventDeleteVersionDialog(t, f)

	// After the final one is deleted, it should no longer be UpdateListingDialog, but should return to the list page
	// This is confirmed by PushState not nil inside this method
	flowDeleteVersion_Step33_Event_publish_eventDeleteVersion(t, f)
	updateExpectedModels(0)

	flowDeleteVersion_Step34_Event___reload__(t, f)
}

func flowDeleteVersion_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-31-v01").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-31-v01"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step01_Event_presets_OpenListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_OpenListingDialog").
		Query("select_id", "1_2024-05-31-v01").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsListingDialog = true }, 100)", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step02_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-26-v06").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-31-v01").
		Query("version_name", "2024-05-26-x06").
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

func flowDeleteVersion_Step03_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-26-v06").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-31-v01").
		Query("version_name", "2024-05-26-x06").
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
	assert.Equal(t, "plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-31-v01\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step04_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-31-v01").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step05_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v04").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step06_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-26-v04").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v04").
		Query("version_name", "2024-05-26-x04").
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

func flowDeleteVersion_Step07_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-26-v04").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v04").
		Query("version_name", "2024-05-26-x04").
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
	assert.Equal(t, "plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-31-v01\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step08_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-31-v01").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step09_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v03").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step10_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-31-v01").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v03").
		Query("version_name", "2024-05-31-v01").
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

func flowDeleteVersion_Step11_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-31-v01").
		Query("id", "1_2024-05-31-v01").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v03").
		Query("version_name", "2024-05-31-v01").
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
	assert.Equal(t, "vars.publish_VarCurrentDisplayID = \"1_2024-05-26-v05\"; vars.presetsRightDrawer = false; plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \"1_2024-05-26-v05\").go(); plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-26-v03\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step12_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-26-v05").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-26-v05"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step13_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v03").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step14_Event_publish_eventSelectVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("publish_eventSelectVersion").
		Query("select_id", "1_2024-05-26-v03").
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
	assert.Equal(t, "vars.presetsListingDialog = false; if (!!vars.publish_VarCurrentDisplayID && vars.publish_VarCurrentDisplayID != \"1_2024-05-26-v03\") { vars.presetsRightDrawer = false;plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \"1_2024-05-26-v03\").go() }", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step15_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-26-v03").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-26-v03"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step16_Event_presets_OpenListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_OpenListingDialog").
		Query("select_id", "1_2024-05-26-v03").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsListingDialog = true }, 100)", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step17_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-26-v03").
		Query("id", "1_2024-05-26-v03").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v03").
		Query("version_name", "2024-05-26-v03").
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

func flowDeleteVersion_Step18_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-26-v03").
		Query("id", "1_2024-05-26-v03").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v03").
		Query("version_name", "2024-05-26-v03").
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
	assert.Equal(t, "vars.publish_VarCurrentDisplayID = \"1_2024-05-26-v02\"; vars.presetsRightDrawer = false; plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \"1_2024-05-26-v02\").go(); plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-26-v02\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step19_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-26-v02").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-26-v02"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step20_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v02").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step21_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v01").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step22_Event_publish_eventSelectVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("publish_eventSelectVersion").
		Query("select_id", "1_2024-05-26-v01").
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
	assert.Equal(t, "vars.presetsListingDialog = false; if (!!vars.publish_VarCurrentDisplayID && vars.publish_VarCurrentDisplayID != \"1_2024-05-26-v01\") { vars.presetsRightDrawer = false;plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \"1_2024-05-26-v01\").go() }", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step23_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-26-v01").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-26-v01"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step24_Event_presets_OpenListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_OpenListingDialog").
		Query("select_id", "1_2024-05-26-v01").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsListingDialog = true }, 100)", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step25_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-26-v01").
		Query("id", "1_2024-05-26-v01").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v01").
		Query("version_name", "2024-05-26-v01").
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

func flowDeleteVersion_Step26_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-26-v01").
		Query("id", "1_2024-05-26-v01").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v01").
		Query("version_name", "2024-05-26-v01").
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
	assert.Equal(t, "vars.publish_VarCurrentDisplayID = \"1_2024-05-26-v05\"; vars.presetsRightDrawer = false; plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \"1_2024-05-26-v05\").go(); plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-26-v05\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step27_Event_presets_DetailingDrawer(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", "1_2024-05-26-v05").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct 1_2024-05-26-v05"),
	)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step28_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v05").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step29_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-26-v05").
		Query("id", "1_2024-05-26-v02").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v05").
		Query("version_name", "2024-05-26-v02").
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

func flowDeleteVersion_Step30_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-26-v05").
		Query("id", "1_2024-05-26-v02").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v05").
		Query("version_name", "2024-05-26-v02").
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
	assert.Equal(t, "plaid().vars(vars).locals(locals).form(form).url(\"/samples/publish-example/with-publish-products-version-list-dialog\").queries({\"select_id\":[\"1_2024-05-26-v05\"]}).eventFunc(\"presets_UpdateListingDialog\").go()", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step31_Event_presets_UpdateListingDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("presets_UpdateListingDialog").
		Query("select_id", "1_2024-05-26-v05").
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
	assert.Equal(t, "listingDialogContentPortal", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "\nvar listingDialogElem = document.getElementById('listingDialog'); \nif (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {\n    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';\n};", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step32_Event_publish_eventDeleteVersionDialog(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersionDialog").
		Query("current_display_id", "1_2024-05-26-v05").
		Query("id", "1_2024-05-26-v05").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v05").
		Query("version_name", "2024-05-26-x05").
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

func flowDeleteVersion_Step33_Event_publish_eventDeleteVersion(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products-version-list-dialog").
		EventFunc("publish_eventDeleteVersion").
		Query("current_display_id", "1_2024-05-26-v05").
		Query("id", "1_2024-05-26-v05").
		Query("overlay", "dialog").
		Query("presets_listing_queries", "select_id=1_2024-05-26-v05").
		Query("version_name", "2024-05-26-x05").
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.PageTitle)
	assert.False(t, resp.Reload)
	assert.NotNil(t, resp.PushState)
	assert.False(t, resp.PushState.MyMergeQuery)
	assert.Equal(t, "/samples/publish-example/with-publish-products", resp.PushState.MyURL)
	assert.Empty(t, resp.PushState.MyStringQuery)
	assert.Empty(t, resp.PushState.MyClearMergeQueryKeys)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDeleteVersion_Step34_Event___reload__(t *testing.T, f *FlowDeleteVersion) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("__reload__").
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Listing WithPublishProducts - Admin", resp.PageTitle)
	assert.True(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}
