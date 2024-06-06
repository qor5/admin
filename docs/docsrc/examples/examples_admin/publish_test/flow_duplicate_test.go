package publish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/docs/v3/docsrc/examples/examples_admin"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
)

var dataSeedForFlowDuplicate = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('1', '2024-05-26 13:12:06.408234+00', '2024-05-26 13:12:06.408234+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-05-26-v01', '2024-05-26-v01', '');
`, []string{"with_publish_products"}))

type FlowDuplicate struct {
	*Flow

	// params
	ID string

	// flow vars
	DuplicateID string
}

func TestFlowDuplicate(t *testing.T) {
	dataSeedForFlowDuplicate.TruncatePut(SQLDB)
	flowDuplicate(t, &FlowDuplicate{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},
		ID: "1_2024-05-26-v01",
	})
}

func flowDuplicate(t *testing.T, f *FlowDuplicate) {
	// Add a new resource to test whether the current case will be affected
	flowNew(t, &FlowNew{
		Flow:  f.Flow,
		Name:  "TheTroublemakerProduct",
		Price: 1031,
	})

	oid, over := MustIDVersion(f.ID)

	// ensure old exists
	var from examples_admin.WithPublishProduct
	require.NoError(t, f.db.Where("id = ? AND version = ?", oid, over).First(&from).Error)

	flowDuplicate_Step00_Event_presets_DetailingDrawer(t, f).Then(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
		assert.True(t, ContainsVersionBar(w.Body.String()))
	})

	// ensure new not exists
	nextVersion, err := GetNextVersion(f.ID)
	assert.NoError(t, err)
	f.DuplicateID = nextVersion

	nid, nver := MustIDVersion(f.DuplicateID)
	assert.ErrorIs(t, f.db.Where("id = ? AND version = ?", nid, nver).First(&examples_admin.WithPublishProduct{}).Error, gorm.ErrRecordNotFound)

	// RunScript inside ensures its interaction: reload.then(openDrawer)
	flowDuplicate_Step01_Event_publish_EventDuplicateVersion(t, f)

	// ensure new exists
	var m examples_admin.WithPublishProduct
	require.NoError(t, f.db.Where("id = ? AND version = ?", nid, nver).First(&m).Error)

	// for compare, change from to as expected
	from.Model = gorm.Model{}
	from.Status = publish.Status{Status: publish.StatusDraft}
	from.Schedule = publish.Schedule{}
	from.Version = publish.Version{
		Version:       nver,
		VersionName:   nver,
		ParentVersion: from.Version.Version,
	}
	m.Model = gorm.Model{}
	assert.Equal(t, from, m)

	// ensure the list reload first
	flowDuplicate_Step02_Event___reload__(t, f)

	// ensure it can be opened
	flowDuplicate_Step03_Event_presets_DetailingDrawer(t, f)
}

func flowDuplicate_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowDuplicate) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
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
	assert.Equal(t, "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)", resp.RunScript)

	testflow.Validate(t, w, r,
		testflow.OpenRightDrawer("WithPublishProduct "+f.ID),
	)
	return testflow.NewThen(t, w, r)
}

func flowDuplicate_Step01_Event_publish_EventDuplicateVersion(t *testing.T, f *FlowDuplicate) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("publish_EventDuplicateVersion").
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
	assert.Empty(t, resp.UpdatePortals)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "plaid().vars(vars).locals(locals).form(form).go().then(function(r){ vars.presetsListingDialog = false; vars.presetsRightDrawer = false; plaid().vars(vars).locals(locals).form(form).eventFunc(\"presets_DetailingDrawer\").query(\"id\", \""+f.DuplicateID+"\").go(); vars.presetsMessage = { show: true, message: \"Successfully Created\", color: \"success\"} })", resp.RunScript)

	return testflow.NewThen(t, w, r)
}

func flowDuplicate_Step02_Event___reload__(t *testing.T, f *FlowDuplicate) *testflow.Then {
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

func flowDuplicate_Step03_Event_presets_DetailingDrawer(t *testing.T, f *FlowDuplicate) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/samples/publish-example/with-publish-products").
		EventFunc("presets_DetailingDrawer").
		Query("id", f.DuplicateID).
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
		testflow.OpenRightDrawer("WithPublishProduct "+f.DuplicateID),
	)

	return testflow.NewThen(t, w, r)
}
