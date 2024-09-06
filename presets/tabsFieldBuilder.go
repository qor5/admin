package presets

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type TabsFieldBuilder struct {
	FieldBuilder
	TabName           []string
	TabComponentFuncs []FieldComponentFunc
}

func (tb *TabsFieldBuilder) appendTabField(tabName string, comp FieldComponentFunc) {
	tb.TabName = append(tb.TabName, tabName)
	tb.TabComponentFuncs = append(tb.TabComponentFuncs, comp)
}

func (tb *TabsFieldBuilder) ComponentFunc() FieldComponentFunc {
	return func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var tabs, contents h.HTMLComponents
		for _, name := range tb.TabName {
			tabs = append(tabs, v.VTab(h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, name))).Value(name))
		}
		for i, comp := range tb.TabComponentFuncs {
			contents = append(contents,
				v.VTabsWindowItem(
					comp(obj, field, ctx),
				).Value(tb.TabName[i]))
		}
		defaultTab := tb.TabName[0]
		return web.Scope(
			v.VTabs(
				// v.VTab(h.Text(msgr.FormTitle)).Value("default"),
				h.Components(tabs...),
			).Class("v-tabs--fixed-tabs").Attr("v-model", "locals.tab"),

			v.VTabsWindow(
				contents...,
			).Attr("v-model", "locals.tab"),
		).VSlot("{ locals }").Init(fmt.Sprintf(`{tab: '%s'}`, defaultTab))
	}
}
