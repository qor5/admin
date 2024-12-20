package examples_admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func LinkageSelectFilterItemRemoteExample(b *presets.Builder, mux examples.Muxer, db *gorm.DB) http.Handler {
	b.DataOperator(gorm2op.DataOperator(db))
	labels := []string{"Province", "City", "District"}
	_ = db.AutoMigrate(&examples_presets.Address{})
	mb := b.Model(&examples_presets.Address{})

	eb := mb.Editing("ProvinceCityDistrict")

	remoteUrl := "/examples/api/linkage-select-server"
	eb.Field("ProvinceCityDistrict").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*examples_presets.Address)
		if field.Errors == nil {
			field.Errors = []string{}
		}
		return vx.VXLinkageSelectRemote().
			Attr(presets.VFieldError(field.Name,
				[]*Item{
					getItem(p.Province),
					getItem(p.City),
					getItem(p.District),
				}, field.Errors)...).
			Labels(labels...).
			RemoteUrl(remoteUrl).
			IsPaging(true).
			LevelStart(1)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		var vs []string
		vErr := &web.ValidationErrors{}
		hasErr := false
		for i := 0; i < 3; i++ {
			val := ctx.Param(fmt.Sprintf("%s[%v].Name", field.FormKey, i))
			vs = append(vs, val)
			msg := ""
			if val == "" {
				hasErr = true
				msg = "Not Empty"
			}
			vErr.FieldError(field.FormKey, msg)
		}

		m := obj.(*examples_presets.Address)
		m.Province = vs[0]
		m.City = vs[1]
		m.District = vs[2]
		if hasErr {
			return vErr
		}
		return
	})

	lb := mb.Listing()
	options := vx.DefaultVXLinkageSelectRemoteOptions("/examples/api/linkage-select-server")
	wrapInputs := make([]func(val string) interface{}, 0)
	for i := 0; i < 3; i++ {
		wrapInputs = append(wrapInputs, func(val string) interface{} {
			return strings.Split(val, options.Separator)[0]
		})
	}

	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:      "province_city_district",
				Label:    "Province&City&District",
				ItemType: vx.ItemTypeLinkageSelectRemote,
				ValuesAre: []*Item{
					getItem("浙江"),
					getItem("杭州"),
					getItem("拱墅区"),
				},
				LinkageSelectData: vx.FilterLinkageSelectData{
					Labels:                     labels,
					SelectOutOfOrder:           true,
					SQLConditions:              []string{"province = ?", "city = ?", "district = ?"},
					LinkageSelectRemoteOptions: options,
					WrapInput:                  wrapInputs,
				},
			},
		}
	})
	sever := LinkageSelectFilterItemRemoteServer{}
	mux.Handle(remoteUrl, &sever)
	return b
}
