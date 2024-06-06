package examples_presets

// @snippet_begin(LinkageSelectFilterItem)
import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func PresetsLinkageSelectFilterItem(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&Address{})

	eb := mb.Editing("ProvinceCityDistrict")

	eb.Field("ProvinceCityDistrict").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		m := obj.(*Address)
		return vx.VXLinkageSelect().
			Attr(web.VField(field.Name, []string{m.Province, m.City, m.District})...).
			Items(getLinkageProvinceCityDistrictItems()...).
			Labels(getLinkageProvinceCityDistrictLabels()...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		vs := ctx.R.Form["ProvinceCityDistrict"]
		m := obj.(*Address)
		m.Province = vs[0]
		m.City = vs[1]
		m.District = vs[2]
		return nil
	})

	lb := mb.Listing()

	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:      "province_city_district",
				Label:    "Province&City&District",
				ItemType: vx.ItemTypeLinkageSelect,
				LinkageSelectData: vx.FilterLinkageSelectData{
					Items:            getLinkageProvinceCityDistrictItems(),
					Labels:           getLinkageProvinceCityDistrictLabels(),
					SelectOutOfOrder: false,
					SQLConditions:    []string{"province = ?", "city = ?", "district = ?"},
				},
				ValuesAre: []string{},
			},
		}
	})
	return
}

func getLinkageProvinceCityDistrictLabels() []string {
	return []string{"Province", "City", "District"}
}

func getLinkageProvinceCityDistrictItems() [][]*vx.LinkageSelectItem {
	return [][]*vx.LinkageSelectItem{
		{
			// use ID as Name if Name is empty
			{ID: "浙江", ChildrenIDs: []string{"杭州", "宁波"}},
			{ID: "江苏", ChildrenIDs: []string{"南京", "苏州"}},
		},
		{
			{ID: "杭州", ChildrenIDs: []string{"拱墅区", "西湖区"}},
			{ID: "宁波", ChildrenIDs: []string{"镇海区", "鄞州区"}},
			{ID: "南京", ChildrenIDs: []string{"鼓楼区", "玄武区"}},
			{ID: "苏州", ChildrenIDs: []string{"常熟区", "吴江区"}},
		},
		{
			{ID: "拱墅区"},
			{ID: "西湖区"},
			{ID: "镇海区"},
			{ID: "鄞州区"},
			{ID: "鼓楼区"},
			{ID: "玄武区"},
			{ID: "常熟区"},
			{ID: "吴江区"},
		},
	}
}

// @snippet_end
