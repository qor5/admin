package examples_vuetifyx

// @snippet_begin(VuetifyComponentsLinkageSelect)

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/theplant/htmlgo"
)

func VuetifyComponentsLinkageSelect(ctx *web.EventContext) (pr web.PageResponse, err error) {
	labels := []string{
		"Province",
		"City",
		"District",
	}
	items := [][]*vx.LinkageSelectItem{
		{
			{ID: "1", Name: "浙江", ChildrenIDs: []string{"1", "2"}},
			{ID: "2", Name: "江苏", ChildrenIDs: []string{"3", "4"}},
		},
		{
			{ID: "1", Name: "杭州", ChildrenIDs: []string{"1", "2"}},
			{ID: "2", Name: "宁波", ChildrenIDs: []string{"3", "4"}},
			{ID: "3", Name: "南京", ChildrenIDs: []string{"5", "6"}},
			{ID: "4", Name: "苏州", ChildrenIDs: []string{"7", "8"}},
		},
		{
			{ID: "1", Name: "拱墅区"},
			{ID: "2", Name: "西湖区"},
			{ID: "3", Name: "镇海区"},
			{ID: "4", Name: "鄞州区"},
			{ID: "5", Name: "鼓楼区"},
			{ID: "6", Name: "玄武区"},
			{ID: "7", Name: "常熟区"},
			{ID: "8", Name: "吴江区"},
		},
	}

	pr.Body = VContainer(
		htmlgo.H3("Basic"),
		vx.VXLinkageSelect().Items(items...).Labels(labels...),
		htmlgo.H3("SelectOutOfOrder"),
		vx.VXLinkageSelect().Items(items...).Labels(labels...).SelectOutOfOrder(true),
		htmlgo.H3("Chips"),
		vx.VXLinkageSelect().Items(items...).Labels(labels...).Chips(true),
		htmlgo.H3("Row"),
		vx.VXLinkageSelect().Items(items...).Labels(labels...).Row(true),
	)

	return pr, nil
}

var VuetifyComponentsLinkageSelectPB = web.Page(VuetifyComponentsLinkageSelect)

var VuetifyComponentsLinkageSelectPath = examples.URLPathByFunc(VuetifyComponentsLinkageSelect)

// @snippet_end
