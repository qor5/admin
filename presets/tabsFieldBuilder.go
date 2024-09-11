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

	tabsOrderFunc func(obj interface{}, field *FieldContext, ctx *web.EventContext) []string

	TabName           []string
	TabComponentFuncs []FieldComponentFunc
}

type TabFieldBuilder struct {
	TabName           string
	TabComponentFuncs FieldComponentFunc
}

func NewTabsFieldBuilder() *TabsFieldBuilder {
	return &TabsFieldBuilder{}
}

func (tb *TabsFieldBuilder) TabsOrderFunc(v func(obj interface{}, field *FieldContext, ctx *web.EventContext) []string) *TabsFieldBuilder {
	if v == nil {
		panic("value required")
	}
	tb.tabsOrderFunc = v
	return tb
}

func (tb *TabsFieldBuilder) AppendTabField(tabName string, comp FieldComponentFunc) {
	tb.TabName = append(tb.TabName, tabName)
	tb.TabComponentFuncs = append(tb.TabComponentFuncs, comp)
}

func (tb *TabsFieldBuilder) ComponentFunc() FieldComponentFunc {
	return func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var tabs, contents h.HTMLComponents
		if tb.tabsOrderFunc == nil {
			for i, name := range tb.TabName {
				tabs = append(tabs, v.VTab(h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, name))).Value(name))
				contents = append(contents,
					v.VTabsWindowItem(
						tb.TabComponentFuncs[i](obj, field, ctx),
					).Value(tb.TabName[i]))
			}
		} else {
			tabsOrder := tb.tabsOrderFunc(obj, field, ctx)
			for _, tab := range tabsOrder {
				for i, name := range tb.TabName {
					if name == tab {
						tabs = append(tabs, v.VTab(h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, name))).Value(name))
						contents = append(contents,
							v.VTabsWindowItem(
								tb.TabComponentFuncs[i](obj, field, ctx),
							).Value(tb.TabName[i]))
					}
				}
			}
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
