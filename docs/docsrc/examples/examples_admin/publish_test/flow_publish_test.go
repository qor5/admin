package publish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/utils/testflow"
)

var dataSeedForFlowPublish = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 10:03:29.389412+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v01', '2024-07-04-v01', '');
`, []string{"with_publish_products"}))

type FlowPublish struct {
	*Flow

	// params
	ID string
}

func TestFlowPublish(t *testing.T) {
	dataSeedForFlowPublish.TruncatePut(SQLDB)

	flowPublish(t, &FlowPublish{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},
		ID: "32_2024-07-04-v01",
	})
}

func flowPublish(t *testing.T, f *FlowPublish) {
	// Add a new resource to test whether the current case will be affected
	flowNew(t, &FlowNew{
		Flow:  f.Flow,
		Name:  "TheTroublemakerProduct",
		Price: 1031,
	})

	id, ver := MustIDVersion(f.ID)
	db := f.db.Where("id = ? AND version = ?", id, ver)

	var prev examples_admin.WithPublishProduct
	require.NoError(t, db.First(&prev).Error)

	ensureVersionBarDisplay := func(btnPublish, btnsAfterPublish, btnSchedule bool) testflow.ValidatorFunc {
		return testflow.Combine(
			testflow.WrapEvent(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, e multipartestutils.TestEventResponse) {
				// Ensure Publish button is displayed
				assert.Equal(t, btnPublish, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Publish</v-btn>"), "btnPublish display")
				// Ensure Unpublish and Republish buttons are displayed
				assert.Equal(t, btnsAfterPublish, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Unpublish</v-btn>", ">Republish</v-btn>"), "btnsAfterPublish display")
				// Ensure the schedule button is displayed
				assert.Equal(t, btnSchedule, testflow.ContainsInOrder(e.UpdatePortals[0].Body, "publish_eventSchedulePublishDialog"), "btnSchedule display")
			}),
		)
	}

	// Open the drawer and confirm the display of draft status
	flowPublish_Step00_Event_presets_DetailingDrawer(t, f).ThenValidate(ensureVersionBarDisplay(true, false, true))

	// Execute publish and ensure data is modified, internally ensures Reload
	previous := f.db.NowFunc()
	assertAfterPublish := func() {
		var m examples_admin.WithPublishProduct
		require.NoError(t, db.First(&m).Error)
		assert.True(t, m.ActualStartAt.After(previous))
		assert.True(t, m.ActualEndAt == nil)
		assert.True(t, m.ScheduledStartAt == nil)
		prev = m
	}
	flowPublish_Step01_Event_publish_EventPublish(t, f)
	assertAfterPublish()

	// Open the drawer and confirm the display of online status
	flowPublish_Step02_Event_presets_DetailingDrawer(t, f).ThenValidate(ensureVersionBarDisplay(false, true, true))

	previous = f.db.NowFunc()
	flowPublish_Step03_Event_publish_EventRepublish(t, f)
	assertAfterPublish()

	// Open the drawer and confirm the display after republishing
	flowPublish_Step04_Event_presets_DetailingDrawer(t, f).ThenValidate(ensureVersionBarDisplay(false, true, true))

	previous = f.db.NowFunc()
	flowPublish_Step05_Event_publish_EventUnpublish(t, f)
	{
		var m examples_admin.WithPublishProduct
		require.NoError(t, db.First(&m).Error)
		assert.True(t, m.ActualEndAt.After(previous))
		assert.True(t, m.ScheduledStartAt == nil)
		assert.True(t, m.ScheduledEndAt == nil)
		prev = m
	}

	// Open the drawer and confirm the display of offline status
	flowPublish_Step06_Event_presets_DetailingDrawer(t, f).ThenValidate(ensureVersionBarDisplay(true, false, true))
}

func flowPublish_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", f.ID).
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
		testflow.OpenRightDrawer("WithPublishProduct "+f.ID),
	)

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step01_Event_publish_EventPublish(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_EventPublish").
		Query("id", f.ID).
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// assert.Equal(t, "Listing WithPublishProducts - Admin", resp.PageTitle)
	// assert.True(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Equal(t, testflow.RemoveTime(`plaid().vars(vars).locals(locals).form(form).dash(dash).mergeQuery(true).go().then(function(r){ vars.presetsMessage = { show: true, message: "Successfully Publish", color: "success"} })`), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step02_Event_presets_DetailingDrawer(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", f.ID).
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
		testflow.OpenRightDrawer("WithPublishProduct "+f.ID),
	)

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step03_Event_publish_EventRepublish(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_EventRepublish").
		Query("id", f.ID).
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// assert.Equal(t, "Listing WithPublishProducts - Admin", resp.PageTitle)
	// assert.True(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Equal(t, testflow.RemoveTime(`plaid().vars(vars).locals(locals).form(form).dash(dash).mergeQuery(true).go().then(function(r){ vars.presetsMessage = { show: true, message: "Successfully Publish", color: "success"} })`), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step04_Event_presets_DetailingDrawer(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", f.ID).
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
		testflow.OpenRightDrawer("WithPublishProduct "+f.ID),
	)

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step05_Event_publish_EventUnpublish(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_EventUnpublish").
		Query("id", f.ID).
		BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// assert.Equal(t, "Listing WithPublishProducts - Admin", resp.PageTitle)
	// assert.True(t, resp.Reload)
	assert.Nil(t, resp.PushState)
	assert.Empty(t, resp.RedirectURL)
	assert.Empty(t, resp.ReloadPortals)
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Equal(t, testflow.RemoveTime(`plaid().vars(vars).locals(locals).form(form).dash(dash).mergeQuery(true).go().then(function(r){ vars.presetsMessage = { show: true, message: "Successfully Unpublish", color: "success"} })`), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}

func flowPublish_Step06_Event_presets_DetailingDrawer(t *testing.T, f *FlowPublish) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", f.ID).
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
		testflow.OpenRightDrawer("WithPublishProduct "+f.ID),
	)

	return testflow.NewThen(t, w, r)
}
