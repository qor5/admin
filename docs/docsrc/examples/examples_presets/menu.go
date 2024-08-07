package examples_presets

import (
	"net/http"
	"net/url"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type (
	music struct{}
	video struct{}
	book  struct{}
)

func PresetsOrderMenu(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.Model(&music{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("music"),
			).Class("text-center mt-8"),
		)
		return
	})
	b.Model(&video{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("video"),
			).Class("text-center mt-8"),
		)
		return
	})
	b.Model(&book{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("book"),
			).Class("text-center mt-8"),
		)
		return
	})
	// @snippet_begin(MenuOrderSample)
	b.MenuOrder(
		"books",
		"videos",
		"musics",
	)
	// @snippet_end
	return
}

func PresetsCustomizeMenu(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.Model(&music{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("music"),
			).Class("text-center mt-8"),
		)
		return
	})
	b.Model(&video{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("video"),
			).Class("text-center mt-8"),
		)
		return
	})
	mb = b.Model(&book{}).MenuIcon("mdi-book")
	// @snippet_begin(MenuCustomizeSample)
	mb.DefaultURLQueryFunc(func(r *http.Request) url.Values {
		return url.Values{
			"extra": []string{"abc"},
		}
	})
	mb.MenuItem(mb.DefaultMenuItem(func(evCtx *web.EventContext, isSub bool, menuIcon string, children ...h.HTMLComponent) ([]h.HTMLComponent, error) {
		return []h.HTMLComponent{
			h.Iff(menuIcon != "", func() h.HTMLComponent {
				return web.Slot(vuetify.VIcon(menuIcon)).Name(vuetify.VSlotPrepend)
			}),
			vuetify.VListItemTitle().Class("d-flex align-center").Children(
				h.Text(mb.Info().LabelName(evCtx, false)),
				vuetify.VSpacer(),
				vuetify.VBadge().Class("pe-1").Content(1).Dot(true).Color(vuetify.ColorError).Children(
					vuetify.VIcon("mdi-bell-outline").Size(20).Color("grey-darken-1"),
				),
			),
		}, nil
	}))
	// @snippet_end
	mb.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("book"),
			).Class("text-center mt-8"),
		)
		return
	})

	b.MenuOrder(
		"books",
		b.MenuGroup("Media").SubItems(
			"videos",
			"musics",
		).Icon("mdi-video"),
	)
	return
}

func PresetsGroupMenu(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.Model(&music{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("music"),
			).Class("text-center mt-8"),
		)
		return
	})
	b.Model(&video{}).Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("video"),
			).Class("text-center mt-8"),
		)
		return
	})
	// @snippet_begin(MenuGroupSample)
	mb = b.Model(&book{}).MenuIcon("mdi-book")

	mb.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = vuetify.VContainer(
			h.Div(
				h.H1("book"),
			).Class("text-center mt-8"),
		)
		return
	})

	b.MenuOrder(
		"books",
		b.MenuGroup("Media").SubItems(
			"videos",
			"musics",
		).Icon("mdi-video"),
	)
	// @snippet_end
	return
}
