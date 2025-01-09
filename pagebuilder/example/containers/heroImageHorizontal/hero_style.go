package heroImageHorizontal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers/tailwind"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
)

type heroStyle struct {
	Layout          string
	TopSpace        int
	BottomSpace     int
	ImgInitial      bool
	ImageBackground media_library.MediaBox `sql:"type:text;"`
}

func (this heroStyle) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *heroStyle) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetHeroStyleComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&heroStyle{}).Only("Layout", "TopSpace", "BottomSpace", "ImageBackground")

	fb.Field("Layout").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		// return presets.SelectField(obj, field, ctx).Items([]string{"left", "right"})

		return Div(
			vx.VXLabel(Text(field.Label)).Class("mb-2"),
			v.VItemGroup(
				v.VItem(
					v.VBtn("").Children(
						Div(
							Div(
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 40px; height: 2px; background: #bdbdbd"),
							).Class("d-flex justify-space-between flex-column").Style("height: 24px"),
							Div().Class("ml-2").Style("background: #bdbdbd;width: 24px;height: 24px;border-radius: 2px;")).Class("d-flex justify-space-between"),
					).
						Class("mr-2 px-2").
						Height(40).
						Attr("@click", "toggle").
						Attr(":class", "['d-flex', 'align-center']").
						Attr(":color", "isSelected ? 'primary' : 'grey'").
						Attr(":style", "{background: isSelected ? '#E6EDFE' : ''}").
						Attr(":variant", "isSelected ? 'outlined' : 'tonal'"),
				).
					Value("left").
					Attr("v-slot", "{ isSelected, toggle }"),
				v.VItem(
					v.VBtn("").Children(
						Div(
							Div().Class("mr-2").Style("background: #bdbdbd;width: 24px;height: 24px;border-radius: 2px;"),
							Div(
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 72px; height: 2px; background: #bdbdbd"),
								Div().Style("width: 40px; height: 2px; background: #bdbdbd"),
							).Class("d-flex justify-space-between flex-column").Style("height: 24px"),
						).Class("d-flex justify-space-between"),
					).
						Class("mr-2 px-2").
						Height(40).
						Attr("@click", "toggle").
						Attr(":class", "['d-flex', 'align-center']").
						Attr(":color", "isSelected ? 'primary' : 'grey'").
						Attr(":style", "{background: isSelected ? '#E6EDFE' : ''}").
						Attr(":variant", "isSelected ? 'outlined' : 'tonal'"),
				).Value("right").Attr("v-slot", "{ isSelected, toggle }"),
			).Class("d-flex").
				Attr(":mandatory", "true").
				// "Style.Layout"
				Attr(web.VField(field.FormKey, reflectutils.MustGet(obj, field.Name))...),
		).Class("mb-4")
	})

	fb.Field("TopSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(tailwind.SpaceOptions)
	})

	fb.Field("BottomSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(tailwind.SpaceOptions)
	})
	// SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("Style").Nested(fb).Label(presets.HiddenLabel)
}
