package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/role"
)

func TestPageBuilderPerm(t *testing.T) {
	h := admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 888},
		Name:  "viwer@theplant.jp",
		Roles: []role.Role{
			{
				Name: models.RoleViewer,
			},
		},
	})
	dbr, _ := TestDB.DB()
	cases := []TestCase{
		{
			Name:  "Page Builder Detail Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				`Page`, "Category", `SEO`, `Activity`,
			},
			ExpectPageBodyNotContains: []string{"v-show='true&&true'"},
		},
		{
			Name:  "Page Builder Detail Editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					Query("containerDataID", "list-content_10_10Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"vx-scroll-iframe"},
			ExpectPageBodyNotContains:     []string{`<v-navigation-drawer :location='"left"'`},
		},
		{
			Name:  "Container Header Update",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headers").
					EventFunc(actions.Update).
					Query(presets.ParamID, "10").
					AddField("Color", "white").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"permission denied"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
