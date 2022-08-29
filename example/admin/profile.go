package admin

import (
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	vx "github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
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
		LayoutConfig(&presets.LayoutConfig{SearchBoxInvisible: true}).
		Label("Profile").MenuIcon("person")

	lb := m.Listing()

	m.RegisterEventFunc("eventSaveProfile", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.Form.Get("id")
		name := ctx.R.Form.Get("name")

		if err = db.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"name": name,
		}).Error; err != nil {
			return
		}

		r.VarsScript = "location.reload();"

		return
	})

	lb.PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		const rowClass = "my-n6"
		u := getCurrentUser(ctx.R)

		var roles []string
		for _, v := range u.Roles {
			roles = append(roles, v.Name)
		}

		r.PageTitle = "Profile"

		r.Body = h.Div(
			VContainer(
				h.Div(
					VRow(
						VCol(
							VTextField().Label("Name").Value(u.Name).FieldName("name"),
						),
					).Class(rowClass),
					VRow(
						VCol(
							vx.VXReadonlyField().Label("Email").Value(u.Account),
						),
					).Class(rowClass),
					VRow(
						VCol(
							vx.VXReadonlyField().Label("Company").Value(u.Company),
						),
					).Class(rowClass),
					VRow(
						VCol(
							vx.VXReadonlyField().Label("Role").Value(strings.Join(roles, ", ")),
						),
					).Class(rowClass),
					VRow(
						VCol(
							vx.VXReadonlyField().Label("Status").Value(u.Status),
						),
					),
					VRow(
						VCol(
							VBtn("").Href("/auth/change-password").
								Outlined(true).Color("primary").
								Children(VIcon("lock_outline").Small(true), h.Text("change password")),
						),
					),
					VRow(
						VCol(
							VBtn("Cancel").Class("mr-1").
								Attr("@click", web.Plaid().Reload().Go()),
							VBtn("Save").Color("primary").
								Attr("@click", web.Plaid().EventFunc("eventSaveProfile").
									Query("id", u.ID).Go(),
								),
						).Class("text-right"),
					),
				).Class("pa-4 my-2"),
			).Fluid(true),
		)
		return
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
