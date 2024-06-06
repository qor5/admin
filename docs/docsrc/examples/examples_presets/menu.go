package examples_presets

import (
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
