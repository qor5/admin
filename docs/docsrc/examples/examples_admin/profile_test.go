package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProfileExample(t *testing.T) {
	currentProfileUser := &ProfileUser{
		Model:   gorm.Model{ID: 1},
		Email:   "admin@theplant.jp",
		Name:    "admin",
		Avatar:  "https://i.pravatar.cc/300",
		Role:    "Admin",
		Status:  "Active",
		Company: "The Plant",
	}

	require.Panics(t, func() {
		pb := presets.New()
		profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
			pb.DisableNotification(true) // .LogoutURL("auth/logout")
		})
	})

	pb := presets.New()
	profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
		pb.DisableNotification(true).LogoutURL("auth/logout")
	})

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "https://i.pravatar.cc/300", "admin", "Active", "admin@theplant.jp", "The Plant", "ADMIN"},
		},
		{
			Name:  "Rename",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*login.ProfileCompo",
	"compo": {
		"id": ""
	},
	"injector": "__profile__",
	"sync_query": false,
	"method": "Rename",
	"request": {
		"name": "adminx"
	}
}
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully renamed"},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				require.Equal(t, "adminx", currentProfileUser.Name)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}
