package examples_presets

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// @snippet_begin(ProfileSample)
func PresetsProfile(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.BrandTitle("Admin").
		ProfileFunc(func(ctx *web.EventContext) h.HTMLComponent {
			// Demo
			name := "QOR5"
			// account := "hello@getqor.com"
			roles := []string{"Developer"}
			return VRow(
				VCol(
					VCard().Class("text-grey-darken-1").Variant("text").Title(name).Subtitle(roles[0]).Children(
						web.Slot(
							VAvatar().Class("ml-1 rounded-lg").Color("blue").Size(48).Children(
								h.Span(string(name[0])).Class("text-white text-h5")),
						).Name("prepend"),
					),
					// VAvatar(
					//	h.Span("QO").Class("text-h5"),
					// ).Attr("color", "blue").Attr("size", "48").Class("rounded-lg"),
					// h.Span(name).Class("text-h6 mx-2 text-grey-darken-1"),
				).Attr("cols", "8"),
				VCol(
					VBtn("").Attr("density", "compact").
						Attr("variant", "text").
						Attr("icon", "mdi-chevron-right").
						Class("text-grey-darken-1"),
				).Attr("cols", "2").Class("d-flex justify-center align-center").Attr("@click", "(e) => {e.view.window.alert('logout')}"),
				VCol(
					VBtn("").Attr("density", "compact").
						Attr("variant", "text").
						Attr("icon", "mdi-bell-outline").
						Class("text-grey-darken-1"),
				).Attr("cols", "2").Class("d-flex align-center justify-center"),
			).Attr("align", "center", "justify", "center")
		})

	// VMenu().Children(
	//	h.Template().Attr("v-slot:activator", "{isActive,props}").Children(
	//		VList(
	//			VListItem(
	//				h.Template().Attr("v-slot:prepend", "{ isActive}").Children(
	//					VAvatar().Class("ml-1").Color("surface-variant").Size(40).Children(
	//						h.Span(string(name[0])).Class("text-white text-h5")),
	//				),
	//			).Title(name).Subtitle(strings.Join(roles, ", ")).Class("pa-0 mb-2"),
	//			VListItem(
	//				h.Template().Attr("v-slot:append", "{ isActive}").Children(
	//					VIcon("mdi-logout").Size("small").Attr("@click", web.Plaid().URL(logoutURL).Go()),
	//				),
	//			).Title(account).Class("pa-0 my-n2 ml-1"),
	//		).Class("pa-0 ma-n4"),
	//	),
	// )
	b.Model(&brand{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = VContainer()
		return
	})
	return
}

// @snippet_end
