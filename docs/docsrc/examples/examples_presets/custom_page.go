package examples_presets

import (
	"fmt"

	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

// @snippet_begin(PresetsCustomPageSample)

func PresetsCustomPage(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsDetailPageTopNotes(b, db)

	b.HandleCustomPage("custom", presets.NewCustomPage(b).Body(func(ctx *web.EventContext) h.HTMLComponent {
		return v.VCard(
			v.VCardItem().Title("New Custom Page"),
		)
	}).Menu(func(ctx *web.EventContext) h.HTMLComponent {
		return nil
	}))
	b.HandleCustomPage("custom/{id}", presets.NewCustomPage(b).Body(func(ctx *web.EventContext) h.HTMLComponent {
		testId := ctx.Param(presets.ParamID)
		return v.VCard(
			v.VCardItem().Title("New Custom Page ID"),
			v.VCardText(
				h.Text(testId)),
		)
	}).Menu(func(ctx *web.EventContext) h.HTMLComponent {
		return nil
	}))
	b.HandleCustomPage("custom-menu", presets.NewCustomPage(b).Body(func(ctx *web.EventContext) h.HTMLComponent {
		return v.VCard(
			v.VCardItem().Title("New Custom Page Show Menu"),
		)
	}))
	dp.Field("NewCustomPage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VBtn("NewCustomPage").Color(v.ColorPrimary).Attr("@click", web.Plaid().PushState(true).URL(b.GetURIPrefix()+"/custom").Go()).Class("mt-2")
	})
	dp.Field("NewCustomPageById").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VBtn("NewCustomPageById").Color(v.ColorPrimary).Attr("@click", web.Plaid().PushState(true).
			URL(fmt.Sprintf(b.GetURIPrefix()+"/custom/%v", ctx.Param(presets.ParamID))).Go()).Class("mt-2")
	})
	dp.Field("NewCustomPageShowMenu").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VBtn("NewCustomPageShowMenu").Color(v.ColorPrimary).Attr("@click", web.Plaid().PushState(true).URL(
			b.GetURIPrefix()+"/custom-menu",
		).Go()).Class("mt-2")
	})
	return
}

// @snippet_end
