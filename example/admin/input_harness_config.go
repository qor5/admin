package admin

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
)

func configInputHarness(b *presets.Builder, db *gorm.DB) {
	harness := b.Model(&models.InputHarness{})

	ed := harness.Editing(
		"TextField1",
		"TextArea1",
		"Switch1",
		"Slider1",
		"Select1",
		//"RangeSlider1",
		"Radio1",
		"FileInput1",
		"Combobox1",
		"Checkbox1",
		"Autocomplete1",
		"ButtonGroup1",
		"ChipGroup1",
		"ItemGroup1",
		"ListItemGroup1",
		"SlideGroup1",
		"ColorPicker1",
		"DatePicker1",
		"DatePickerMonth1",
		"TimePicker1",
	)

	//TextField1       string
	//TextArea1        string
	//Switch1          bool
	//Slider1          int
	//Select1          string
	//RangeSlider1     string
	//Radio1           string
	//FileInput1       string
	//Combobox1        string
	//Checkbox1        string
	//Autocomplete1    string
	//ButtonGroup1     string
	//ChipGroup1       string
	//ItemGroup1       string
	//ListItemGroup1   string
	//SlideGroup1      string
	//ColorPicker1     string
	//DatePicker1      string
	//DatePickerMonth1 string
	//TimePicker1      string

	ed.Field("TextField1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VTextField().FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("TextArea1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VTextarea().FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("Switch1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSwitch().FieldName(field.Name).Label(field.Label).InputValue(field.Value(obj)).Value(field.Value(obj))
		})

	ed.Field("Slider1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSlider().FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("Select1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSelect().FieldName(field.Name).
				Label(field.Label).Value(field.Value(obj)).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"})
		})

	ed.Field("RangeSlider1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VRangeSlider().Attr(web.VFieldName(field.Name)...).
				Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("Radio1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VRadioGroup(
				VRadio().Value("1").Label("Tokyo"),
				VRadio().Value("2").Label("Canberra"),
				VRadio().Value("3").Label("Hangzhou"),
			).Label(field.Label).Value(field.Value(obj)).FieldName(field.Name)
		})

}
