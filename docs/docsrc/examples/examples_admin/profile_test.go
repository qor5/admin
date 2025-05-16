package examples_admin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/stretchr/testify/require"
	h "github.com/theplant/htmlgo"
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
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "https://i.pravatar.cc/300", "admin", "Active", "admin@theplant.jp", "The Plant", "Admin"},
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
		{
			Name:  "Index Page with CustomizeButtons",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb = presets.New()
				profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
					pb.DisableNotification(true).LogoutURL("auth/logout").CustomizeButtons(func(ctx context.Context, buttons ...h.HTMLComponent) ([]h.HTMLComponent, error) {
						return slices.Concat([]h.HTMLComponent{
							v.VBtn("Change Password").Variant(v.VariantTonal).Color(v.ColorSecondary).Attr("@click", "console.log('clicked change password')"),
						}, buttons), nil
					})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "clicked change password", "Change Password", "Logout"},
		},
		{
			Name:  "Index Page with PrependCompos",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb = presets.New()
				profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
					pb.DisableNotification(true).LogoutURL("auth/logout").
						PrependCompos(func(ctx context.Context, profileCompo *plogin.ProfileCompo) ([]h.HTMLComponent, error) {
							profileCompoFromCtx := plogin.ProfileCompoFromContext(ctx)

							return []h.HTMLComponent{
								h.Div().Text("PrependCompos"),
								h.Div().Text(fmt.Sprintf("ProfileCompoEquals: %t", profileCompo == profileCompoFromCtx)),
							}, nil
						})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "PrependCompos", "ProfileCompoEquals: true"},
		},
		{
			Name:  "Index Page with WrapPrependCompo",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb = presets.New()
				profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
					pb.DisableNotification(true).LogoutURL("auth/logout").
						WrapPrependCompo(func(in plogin.CustomizeCompoFunc) plogin.CustomizeCompoFunc {
							return func(ctx context.Context, profileCompo *plogin.ProfileCompo, current h.HTMLComponent) (h.HTMLComponent, error) {
								// current, err := in(ctx, profileCompo, current)
								// if err != nil {
								// 	return nil, err
								// }

								profileCompoFromCtx := plogin.ProfileCompoFromContext(ctx)
								return h.Components(
									h.Div().Text("PrependCompos"),
									h.Div().Text(fmt.Sprintf("ProfileCompoEquals: %t", profileCompo == profileCompoFromCtx)),
								), nil
							}
						})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "PrependCompos", "ProfileCompoEquals: true"},
		},
		{
			Name:  "Index Page with WrapSubtitleCompo",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb = presets.New()
				profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
					pb.DisableNotification(true).LogoutURL("auth/logout").
						WrapSubtitleCompo(func(in plogin.CustomizeCompoFunc) plogin.CustomizeCompoFunc {
							return func(ctx context.Context, profileCompo *plogin.ProfileCompo, current h.HTMLComponent) (h.HTMLComponent, error) {
								// current, err := in(ctx, profileCompo, current)
								// if err != nil {
								// 	return nil, err
								// }

								return h.Div().Class("text-overline text-grey-darken-1 text-truncate").Text("Customize Subtitle"), nil
							}
						})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ProfileCompo", "Customize Subtitle"},
		},
		{
			Name:  "Index Page with WrapRenameCallback",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb = presets.New()
				profileExample(pb, TestDB, currentProfileUser, func(pb *plogin.ProfileBuilder) {
					pb.DisableNotification(true).LogoutURL("auth/logout").
						WrapRenameCallback(func(in plogin.RenameCallback) plogin.RenameCallback {
							return func(ctx context.Context, newName string) error {
								newName += "_suffix"
								return in(ctx, newName)
							}
						})
				})
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
				require.Equal(t, "adminx_suffix", currentProfileUser.Name)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}
