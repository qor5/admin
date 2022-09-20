package admin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	vx "github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/login"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	signOutAllSessionEvent = "signOutAllSessionEvent"
)

func profile(ctx *web.EventContext) h.HTMLComponent {
	u := getCurrentUser(ctx.R)
	if u == nil {
		return VBtn("Login").Text(true).Href("/auth/login")
	}

	var roles []string
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}

	return VMenu().OffsetY(true).Children(
		h.Template().Attr("v-slot:activator", "{on, attrs}").Children(
			VList(
				VListItem(
					VListItemAvatar(
						VAvatar().Class("ml-1").Color("secondary").Size(40).Children(
							h.If(u.OAuthAvatar == "",
								h.Span(getAvatarShortName(u)).Class("white--text text-h5"),
							).Else(
								h.Img(u.OAuthAvatar).Alt(u.Name),
							),
						),
					),
					VListItemContent(
						VListItemTitle(h.Text(u.Name)),
						h.Br(),
						VListItemSubtitle(h.Text(strings.Join(roles, ", "))),
					),
				).Class("pa-0 mb-2"),
				VListItem(
					VListItemContent(
						VListItemTitle(h.Text(u.Account)),
					),
					VListItemIcon(
						VIcon("logout").Small(true).Attr("@click", web.Plaid().URL("/auth/logout").Go()),
					),
				).Class("pa-0 my-n4 ml-1").Dense(true),
			).Class("pa-0 ma-n4"),
		),
	)
}

type Profile struct{}

func configProfile(b *presets.Builder, db *gorm.DB) {
	m := b.Model(&Profile{}).URIName("profile").
		Label("Profile").MenuIcon("person").Singleton(true)

	eb := m.Editing("Info", "Actions")

	m.RegisterEventFunc(signOutAllSessionEvent, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		u := getCurrentUser(ctx.R)
		err = login.SignOutAllOtherSessions(loginBuilder, ctx.W, ctx.R, db, &models.User{}, fmt.Sprint(u.ID), u)
		if err != nil {
			return r, err
		}
		return
	})

	eb.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		u := getCurrentUser(ctx.R)
		if u == nil {
			return nil, errors.New("cannot get current user")
		}
		return u, nil
	})

	eb.SetterFunc(func(obj interface{}, ctx *web.EventContext) {
		u := obj.(*models.User)
		u.Name = ctx.R.FormValue("name")
		return
	})

	eb.Field("Info").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*models.User)
		var roles []string
		for _, v := range u.Roles {
			roles = append(roles, v.Name)
		}

		return h.Div(
			VRow(
				VCol(
					VTextField().Label("Name").Value(u.Name).FieldName("name"),
				),
			).Class("my-n6"),
			VRow(
				VCol(
					vx.VXReadonlyField().Label("Email").Value(u.Account),
				),
			).Class("my-n6"),
			VRow(
				VCol(
					vx.VXReadonlyField().Label("Company").Value(u.Company),
				),
			).Class("my-n6"),
			VRow(
				VCol(
					vx.VXReadonlyField().Label("Role").Value(strings.Join(roles, ", ")),
				),
			).Class("my-n6"),
			VRow(
				VCol(
					vx.VXReadonlyField().Label("Status").Value(u.Status),
				),
			),
		).Class("mt-4 ml-2")
	})

	eb.Field("Actions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var actionBtns h.HTMLComponents

		actionBtns = append(actionBtns,
			VBtn("").Href("/auth/change-password").
				Outlined(true).Color("primary").
				Children(VIcon("lock_outline").Small(true), h.Text("change password")).
				Class("mr-2"),
		)

		actionBtns = append(actionBtns,
			VBtn("").Attr("@click", web.Plaid().EventFunc(signOutAllSessionEvent).Go()).
				Outlined(true).Color("primary").
				Children(VIcon("warning").Small(true), h.Text("Sign out all other sessions")),
		)

		return h.Div(
			actionBtns...,
		).Class("ml-2 mt-8 text-left")
	})
}

func getAvatarShortName(u *models.User) string {
	name := u.Name
	if name == "" {
		name = u.Account
	}
	if rs := []rune(name); len(rs) > 1 {
		name = string(rs[:1])
	}

	return name
}
