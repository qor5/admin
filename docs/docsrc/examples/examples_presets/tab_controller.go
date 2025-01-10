package examples_presets

import (
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

func PresetsEditingTabController(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsHelloWorld(b, db)
	ce.Creating()
	ed := mb.Editing("Notes", "Tabs", "Name", "Email", "Tabs2", "Description", "ApprovedAt")
	ed.Field("Tabs").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		option := presets.TabsControllerOption{
			DefaultIndex: 1,
			Tabs: []presets.TabControllerOption{
				{Tab: v.VTab().Text("t1"), Fields: []string{"Name"}},
				{Tab: v.VTab().Text("t2"), Fields: []string{"Email"}},
			},
			WrapTabComponent: func(comp *vx.VXTabsBuilder) *vx.VXTabsBuilder {
				return comp.UnderlineBorder("full")
			},
		}
		return presets.TabsController(field, &option)
	})
	ed.Field("Tabs2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		option := presets.TabsControllerOption{
			DefaultIndex: 1,
			Tabs: []presets.TabControllerOption{
				{Tab: v.VTab().Text("t3"), Fields: []string{"Description"}},
				{Tab: v.VTab().Text("t4"), Fields: []string{"ApprovedAt"}},
			},
			WrapTabComponent: func(comp *vx.VXTabsBuilder) *vx.VXTabsBuilder {
				return comp.UnderlineBorder("full")
			},
		}
		return presets.TabsController(field, &option)
	})
	return
}
