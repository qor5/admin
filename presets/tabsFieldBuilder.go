package presets

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type TabsFieldBuilder struct {
	FieldBuilder

	tabsOrderFunc func(obj interface{}, field *FieldContext, ctx *web.EventContext) []string

	tabFields []*tabFieldBuilder
}

type tabFieldBuilder struct {
	NameLabel
	tabComponentFuncs FieldComponentFunc
}

func NewTabsFieldBuilder() *TabsFieldBuilder {
	r := &TabsFieldBuilder{}
	r.tabsOrderFunc = func(obj interface{}, field *FieldContext, ctx *web.EventContext) []string {
		var tabOrder []string
		for _, v := range r.tabFields {
			tabOrder = append(tabOrder, v.name)
		}
		return tabOrder
	}
	return r
}

func (tb *TabsFieldBuilder) TabsOrderFunc(v func(obj interface{}, field *FieldContext, ctx *web.EventContext) []string) *TabsFieldBuilder {
	if v == nil {
		panic("value required")
	}
	tb.tabsOrderFunc = v
	return tb
}

func (tb *TabsFieldBuilder) AppendTabField(tabName, tabLabel string, comp FieldComponentFunc) {
	if tabLabel == "" {
		tabLabel = tabName
	}
	field := &tabFieldBuilder{
		NameLabel: NameLabel{
			name:  tabName,
			label: tabLabel,
		},
		tabComponentFuncs: comp,
	}
	tb.tabFields = append(tb.tabFields, field)
}

func (tb *TabsFieldBuilder) ComponentFunc() FieldComponentFunc {
	return func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			tabs, contents h.HTMLComponents
			defaultTab     string
			order          []string
		)
		order = tb.tabsOrderFunc(obj, field, ctx)
		for _, tabName := range order {
			for _, tab := range tb.tabFields {
				if tab.name == tabName {
					tabKey := strcase.ToSnake(tab.name)
					tabs = append(tabs, v.VTab(
						h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, tab.label))).Value(tabKey),
					)
					contents = append(
						contents,
						v.VTabsWindowItem(
							tab.tabComponentFuncs(obj, field, ctx),
						).Value(tabKey),
					)
					if defaultTab == "" {
						defaultTab = tabKey
					}
				}
			}
		}

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
