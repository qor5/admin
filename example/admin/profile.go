package admin

import (
	"github.com/goplaid/web"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/example/models"
	. "github.com/theplant/htmlgo"
)

func profile(ctx *web.EventContext) HTMLComponent {
	u := getCurrentUser(ctx.R)
	if u == nil {
		return VBtn("Login").Text(true).Href("/auth/login")
	}

	return VMenu().OffsetY(true).Children(
		Template().Attr("v-slot:activator", "{on, attrs}").Children(
			VAvatar().Class("ml-4").Color("secondary").Size(40).Attr("v-bind", "attrs").Attr("v-on", "on").Children(
				If(u.AvatarURL == "",
					Span(getAvatarShortName(u)).Class("white--text text-h5"),
				).Else(
					Img(u.AvatarURL).Alt(u.Name),
				),
			),
		),
		VList(
			VListItem(
				VListItemTitle(VBtn("Logout").Text(true).Href("/auth/logout")),
			),
		),
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
