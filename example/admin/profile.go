package admin

import (
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
)

func profile(ctx *web.EventContext) h.HTMLComponent {
	u := getCurrentUser(ctx.R)
	if u == nil {
		return vuetify.VBtn("Login").Text(true).Href("/auth/login")
	}

	var roles []string
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}

	return vuetify.VMenu().OffsetY(true).Children(
		h.Template().Attr("v-slot:activator", "{on, attrs}").Children(
			h.Div(
				vuetify.VRow(
					h.Div(vuetify.VAvatar().Color("primary").Size(24).Children(
						h.If(u.AvatarURL == "",
							vuetify.VIcon("account_circle"),
						).Else(
							h.Img(u.AvatarURL).Alt(u.Name),
						)),
						h.Text(u.Name), h.If(len(u.Roles) > 0, h.Text("("+strings.Join(roles, ",")+")")),
					).Style(`width:100%;`).Class("text-button"),
					h.Div(
						h.Text(u.Email),
					),
				),
			).Attr("v-bind", "attrs").Attr("v-on", "on"),
		),
		vuetify.VList(
			h.Div(
				vuetify.VListItem(
					vuetify.VListItemContent(
						vuetify.VListItemTitle(
							h.Div(h.Text("Logout")).Class("text-button"),
						),
					),
				).Attr("@click", web.Plaid().URL("/auth/logout").Go()),
			),
		).Dense(true),
	)
}

func getAvatarShortName(u *models.User) string {
	name := u.Name
	if name == "" {
		name = u.Email
	}
	if rs := []rune(name); len(rs) > 1 {
		name = string(rs[:1])
	}

	return name
}
