package publish_test

import (
	"encoding/json"
	"fmt"
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
)

var dataSeedForFlowScheduleDraft = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 10:03:29.389412+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v01', '2024-07-04-v01', '');
`, []string{"with_publish_products"}))

var dataSeedForFlowScheduleOnline = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 13:54:50.177453+00', NULL, 'FirstProduct', '123', 'online', '', NULL, NULL, '2024-07-04 13:54:50.171564+00', NULL, '2024-07-04-v01', '2024-07-04-v01', '');
`, []string{"with_publish_products"}))

type FlowSchedule struct {
	*Flow

	ID               string
	ScheduledStartAt time.Time
	ScheduledEndAt   time.Time

	ExpectErrorMessage string
}

func TestFlowSchedule(t *testing.T) {
	cases := []struct {
		desc    string
		startAt time.Time
		endAt   time.Time
		errMsg  string
		online  bool
	}{
		{
			desc:    "empty",
			startAt: time.Time{},
			endAt:   time.Time{},
		},
		{
			desc:    "start < now < end",
			startAt: time.Now().AddDate(0, 0, -1),
			endAt:   time.Now().AddDate(0, 0, 1),
			errMsg:  "Start at should be later than now",
		},
		{
			desc:    "end < now < no start",
			startAt: time.Time{},
			endAt:   time.Now().AddDate(0, 0, -1),
			errMsg:  "End at should be later than now or empty",
		},
		{
			desc:    "now < start == end",
			startAt: time.Now().AddDate(0, 0, 1),
			endAt:   time.Now().AddDate(0, 0, 1),
			errMsg:  "End at should be later than start at",
		},
		{
			desc:    "now < no start < end",
			startAt: time.Time{},
			endAt:   time.Now().AddDate(0, 0, 2),
			errMsg:  "Start at should not be empty",
		},
		{
			desc:    "now < start < end",
			startAt: time.Now().AddDate(0, 0, 2),
			endAt:   time.Now().AddDate(0, 0, 3),
		},
		{
			desc:    "now < start < no end",
			startAt: time.Now().AddDate(0, 0, 2),
			endAt:   time.Time{},
		},
		{
			desc:    "(online) now < no start < end",
			startAt: time.Time{},
			endAt:   time.Now().AddDate(0, 0, 2),
			online:  true,
		},
		{
			desc:    "(online) now < start < end",
			startAt: time.Now().AddDate(0, 0, 1), // start will be ignore
			endAt:   time.Now().AddDate(0, 0, 2),
			online:  true,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if c.online {
				dataSeedForFlowScheduleOnline.TruncatePut(SQLDB)
			} else {
				dataSeedForFlowScheduleDraft.TruncatePut(SQLDB)
			}
			flowSchedule(t, &FlowSchedule{
				Flow:               &Flow{db: DB, h: PresetsBuilder},
				ID:                 "32_2024-07-04-v01",
				ScheduledStartAt:   c.startAt,
				ScheduledEndAt:     c.endAt,
				ExpectErrorMessage: c.errMsg,
			})
		})
	}
}

func scheduledTimeFormat(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04") // minute
}

func flowSchedule(t *testing.T, f *FlowSchedule) {
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

	prevScheduledStartAt := scheduledTimeFormat(prev.ScheduledStartAt)
	prevScheduledEndAt := scheduledTimeFormat(prev.ScheduledEndAt)

	// Ensure the schedule button is displayed; if not, it implies that the state is incorrect
	flowSchedule_Step00_Event_presets_DetailingDrawer(t, f).ThenValidate(testflow.ContainsInOrderAtUpdatePortal(0, "publish_eventSchedulePublishDialog"))

	candidates := []string{
		fmt.Sprintf(`.query("id", %q)`, f.ID),                                // Ensure the correct project is being operated on
		fmt.Sprintf(`[form, {"ScheduledStartAt":%q}]`, prevScheduledStartAt), // Ensure the original planned time is correct
		fmt.Sprintf(`[form, {"ScheduledEndAt":%q}]`, prevScheduledEndAt),
	}
	if prev.Status.Status == publish.StatusOnline {
		candidates = append(candidates[0:1], candidates[2:]...)
	}
	flowSchedule_Step01_Event_publish_eventSchedulePublishDialog(t, f).ThenValidate(
		testflow.ContainsInOrderAtUpdatePortal(0, candidates...),
	)

	if f.ExpectErrorMessage == "" {
		flowSchedule_Step02_Event_publish_eventSchedulePublish(t, f)

		var m examples_admin.WithPublishProduct
		require.NoError(t, db.First(&m).Error)
		if m.Status.Status == publish.StatusOnline {
			assert.Equal(t, "", scheduledTimeFormat(m.ScheduledStartAt))
		} else {
			assert.Equal(t, scheduledTimeFormat(&f.ScheduledStartAt), scheduledTimeFormat(m.ScheduledStartAt))
		}
		assert.Equal(t, scheduledTimeFormat(&f.ScheduledEndAt), scheduledTimeFormat(m.ScheduledEndAt))

		// flowSchedule_Step03_Event___dispatch_stateful_action__(t, f)
		return
	}

	// if error
	flowSchedule_Step05_Event_publish_eventSchedulePublish(t, f)
}

func flowSchedule_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowSchedule) *testflow.Then {
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

func flowSchedule_Step01_Event_publish_eventSchedulePublishDialog(t *testing.T, f *FlowSchedule) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_eventSchedulePublishDialog").
		Query("id", f.ID).
		Query("overlay", "dialog").
		AddField("ScheduledStartAt", scheduledTimeFormat(&f.ScheduledStartAt)).
		AddField("ScheduledEndAt", scheduledTimeFormat(&f.ScheduledEndAt)).
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
	assert.Equal(t, "publish_PortalSchedulePublishDialog", resp.UpdatePortals[0].Name)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowSchedule_Step02_Event_publish_eventSchedulePublish(t *testing.T, f *FlowSchedule) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_eventSchedulePublish").
		Query("id", f.ID).
		AddField("ScheduledStartAt", scheduledTimeFormat(&f.ScheduledStartAt)).
		AddField("ScheduledEndAt", scheduledTimeFormat(&f.ScheduledEndAt)).
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
	// assert.Equal(t, testflow.RemoveTime(`locals.schedulePublishDialog = false; plaid().vars(vars).emit("PresetsNotifModelsUpdatedexamplesAdminWithPublishProduct", {"ids":["32_2024-07-04-v01"],"models":[{"ID":32,"CreatedAt":"2024-07-04T18:03:29.389412+08:00","UpdatedAt":"2024-07-04T19:56:54.972308+08:00","DeletedAt":null,"Name":"FirstProduct","Price":123,"Status":"draft","OnlineUrl":"","ScheduledStartAt":"2024-07-07T00:00:00+08:00","ScheduledEndAt":"2024-07-08T00:00:00+08:00","ActualStartAt":null,"ActualEndAt":null,"Version":"2024-07-04-v01","VersionName":"2024-07-04-v01","ParentVersion":""}]})`), testflow.RemoveTime(resp.RunScript))
	assert.Contains(t, resp.RunScript, `emit("PresetsNotifModelsUpdatedexamplesAdminWithPublishProduct"`)

	return testflow.NewThen(t, w, r)
}

func flowSchedule_Step03_Event___dispatch_stateful_action__(t *testing.T, f *FlowSchedule) *testflow.Then {
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

func flowSchedule_Step05_Event_publish_eventSchedulePublish(t *testing.T, f *FlowSchedule) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
		EventFunc("publish_eventSchedulePublish").
		Query("id", f.ID).
		AddField("ScheduledStartAt", scheduledTimeFormat(&f.ScheduledStartAt)).
		AddField("ScheduledEndAt", scheduledTimeFormat(&f.ScheduledEndAt)).
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
	assert.Equal(t, testflow.RemoveTime(fmt.Sprintf("vars.presetsMessage = { show: true, message: \"%s\", color: \"error\"}", f.ExpectErrorMessage)), testflow.RemoveTime(resp.RunScript))

	return testflow.NewThen(t, w, r)
}
