package presets

import (
	"fmt"

	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

func LinkageFieldsController(field *FieldContext, vs ...string) h.HTMLComponent {
	vs = append(vs, field.FormKey)
	return h.Div().Attr("v-on-mounted", fmt.Sprintf(`()=>{
	    dash.__lingkageFields = dash.__lingkageFields??[];
	    dash.__currentValidateKeys = dash.__currentValidateKeys??[];
		dash.__lingkageFields.push(%v)
		if (!vars.__findLinkageFields){
			dash.__findLinkageFields = function findLinkageFields( x) {
    		const result = new Set();
    		dash.__lingkageFields.forEach(subArray => {
        	if (subArray.includes(x)) {
            subArray.forEach(value => {	
			if (value !== x) {
				result.add(value);
				dash.__currentValidateKeys.push(value)
                }
            });
        }
    });
}
}
	}`, h.JSONString(vs)))
}

type (
	TabsControllerOption struct {
		DefaultIndex     int
		Tabs             []TabControllerOption
		WrapTabComponent func(*vx.VXTabsBuilder) *vx.VXTabsBuilder
	}
	TabControllerOption struct {
		Tab    *v.VTabBuilder
		Fields []string
	}
)

func TabsController(field *FieldContext, option *TabsControllerOption) h.HTMLComponent {
	if option == nil || len(option.Tabs) == 0 || option.DefaultIndex >= len(option.Tabs) {
		return nil
	}

	var (
		tabs      []h.HTMLComponent
		tabFields [][]string
		fields    []string
	)
	for index, tabController := range option.Tabs {
		tabs = append(tabs, tabController.Tab.Value(index))
		tabFields = append(tabFields, tabController.Fields)
		fields = append(fields, tabController.Fields...)
	}
	vxTabs := vx.VXTabs(
		tabs...,
	).UnderlineBorder("full")
	if option.WrapTabComponent != nil {
		vxTabs = option.WrapTabComponent(vxTabs)
	}

	return h.Div(
		web.Scope(
			h.Div().Attr("v-on-mounted", fmt.Sprintf(`() => {
dash.visible = dash.visible??{};
dash.visible[%q]=true;
tabsTabsControllerLocals.__setTabFieldVisible = (index) => {
		
		tabsTabsControllerLocals.fields.forEach((fieldFormKey)=>{
			dash.visible[fieldFormKey] = false;
		})
        tabsTabsControllerLocals.items[index].forEach((fieldFormKey) => {
            dash.visible[fieldFormKey] = true;
        })}
   tabsTabsControllerLocals.__setTabFieldVisible(tabsTabsControllerLocals.tab)
}`, field.FormKey)).Style("display:none"),
			vxTabs.Attr("@update:model-value",
				`
tabsTabsControllerLocals.__setTabFieldVisible($event)`).
				Attr("v-model", "tabsTabsControllerLocals.tab"),
		).VSlot("{locals:tabsTabsControllerLocals}").
			Init(fmt.Sprintf("{tab:%v,items:%v,fields:%v}", option.DefaultIndex, h.JSONString(tabFields), h.JSONString(fields))),
	).Class("mb-8")
}
