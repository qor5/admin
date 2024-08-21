package publish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
)

var dataSeedForFlowDuplicate = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 10:03:29.389412+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v01', '2024-07-04-v01', '');
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
		ID: "32_2024-07-04-v01",
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
}

func flowDuplicate_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowDuplicate) *testflow.Then {
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

func flowDuplicate_Step01_Event_publish_EventDuplicateVersion(t *testing.T, f *FlowDuplicate) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("/examples/publish-example/with-publish-products").
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
	// assert.Equal(t, testflow.RemoveTime(`locals.commonConfirmDialog = false; plaid().vars(vars).emit("PresetsNotifModelsCreatedexamplesAdminWithPublishProduct", {"models":[{"ID":32,"CreatedAt":"2024-07-04T18:03:54.402772+08:00","UpdatedAt":"2024-07-04T18:03:54.399244+08:00","DeletedAt":null,"Name":"FirstProduct","Price":123,"Status":"draft","OnlineUrl":"","ScheduledStartAt":null,"ScheduledEndAt":null,"ActualStartAt":null,"ActualEndAt":null,"Version":"2024-07-04-v02","VersionName":"2024-07-04-v02","ParentVersion":"2024-07-04-v01"}]}); plaid().vars(vars).emit("PublishNotifVersionSelectedexamplesAdminWithPublishProduct", {"slug":"32_2024-07-04-v02"}); vars.presetsMessage = { show: true, message: "Successfully Created", color: "success"}`), testflow.RemoveTime(resp.RunScript))
	assert.Contains(t, resp.RunScript, `emit("PresetsNotifModelsCreatedexamplesAdminWithPublishProduct"`)

	return testflow.NewThen(t, w, r)
}
