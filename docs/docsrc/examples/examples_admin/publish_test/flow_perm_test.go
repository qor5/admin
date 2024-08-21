package publish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/perm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
)

var dataSeedForFlowPerm = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_publish_products" ("id", "created_at", "updated_at", "deleted_at", "name", "price", "status", "online_url", "scheduled_start_at", "scheduled_end_at", "actual_start_at", "actual_end_at", "version", "version_name", "parent_version") VALUES ('32', '2024-07-04 10:03:29.389412+00', '2024-07-04 10:03:29.389412+00', NULL, 'FirstProduct', '123', 'draft', '', NULL, NULL, NULL, NULL, '2024-07-04-v01', '2024-07-04-v01', '');
`, []string{"with_publish_products"}))

type FlowPerm struct {
	*Flow

	ID string
}

func TestFlowPerm(t *testing.T) {
	dataSeedForFlowPerm.TruncatePut(SQLDB)

	flowPerm(t, &FlowPerm{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},

		ID: "32_2024-07-04-v01",
	})
}

func flowPerm(t *testing.T, f *FlowPerm) {
	type BarDisplay struct {
		BtnDuplicate, BtnPublish, BtnUnpublish, BtnRepublish, BtnSchedule bool
	}

	ensureVersionBarDisplay := func(display BarDisplay) testflow.ValidatorFunc {
		return testflow.Combine(
			testflow.WrapEvent(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, e multipartestutils.TestEventResponse) {
				assert.Equal(t, display.BtnDuplicate, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Duplicate</v-btn>"), "btnDuplicate display")
				assert.Equal(t, display.BtnPublish, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Publish</v-btn>"), "btnPublish display")
				assert.Equal(t, display.BtnUnpublish, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Unpublish</v-btn>"), "btnUnpublish display")
				assert.Equal(t, display.BtnRepublish, testflow.ContainsInOrder(e.UpdatePortals[0].Body, ">Republish</v-btn>"), "btnRepublish display")
				assert.Equal(t, display.BtnSchedule, testflow.ContainsInOrder(e.UpdatePortals[0].Body, "publish_eventSchedulePublishDialog"), "btnSchedule display")
			}),
		)
	}

	cases := []struct {
		desc    string
		ps      []*perm.PolicyBuilder
		display BarDisplay
	}{
		{
			desc: "",
			ps: []*perm.PolicyBuilder{
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermPublish).On(perm.Anything),
			},
			display: BarDisplay{
				BtnDuplicate: true,
				BtnPublish:   false,
				BtnUnpublish: false,
				BtnRepublish: false,
				BtnSchedule:  false,
			},
		},
		{
			desc: "",
			ps: []*perm.PolicyBuilder{
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermDuplicate).On(perm.Anything),
			},
			display: BarDisplay{
				BtnDuplicate: false,
				BtnPublish:   true,
				BtnUnpublish: false,
				BtnRepublish: false,
				BtnSchedule:  true,
			},
		},
		{
			desc: "",
			ps: []*perm.PolicyBuilder{
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermSchedule).On(perm.Anything),
			},
			display: BarDisplay{
				BtnDuplicate: true,
				BtnPublish:   true,
				BtnUnpublish: false,
				BtnRepublish: false,
				BtnSchedule:  false,
			},
		},
		{
			desc: "",
			ps: []*perm.PolicyBuilder{
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermDuplicate).On(perm.Anything),
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermSchedule).On(perm.Anything),
			},
			display: BarDisplay{
				BtnDuplicate: false,
				BtnPublish:   true,
				BtnUnpublish: false,
				BtnRepublish: false,
				BtnSchedule:  false,
			},
		},
		{
			desc: "",
			ps: []*perm.PolicyBuilder{
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermDuplicate).On(perm.Anything),
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(publish.PermPublish).On(perm.Anything),
			},
			display: BarDisplay{
				BtnDuplicate: false,
				BtnPublish:   false,
				BtnUnpublish: false,
				BtnRepublish: false,
				BtnSchedule:  false,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			old := f.h.(*presets.Builder).GetPermission()
			defer func() {
				f.h.(*presets.Builder).Permission(old)
			}()
			f.h.(*presets.Builder).Permission(
				perm.New().Policies(
					append([]*perm.PolicyBuilder{
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
					}, c.ps...)...,
				),
			)
			flowPerm_Step00_Event_presets_DetailingDrawer(t, f).ThenValidate(ensureVersionBarDisplay(c.display))
		})
	}
}

func flowPerm_Step00_Event_presets_DetailingDrawer(t *testing.T, f *FlowPerm) *testflow.Then {
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
