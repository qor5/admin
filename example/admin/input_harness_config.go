package admin

import (
	"fmt"
	"io/ioutil"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
)

func configInputHarness(b *presets.Builder, db *gorm.DB) {
	harness := b.Model(&models.InputHarness{})

	ed := harness.Editing()

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

	//ed.Field("RangeSlider1").
	//	ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	//		return VRangeSlider().Attr(web.VFieldName(field.Name)...).
	//			Label(field.Label).Value(field.Value(obj))
	//	})

	ed.Field("Radio1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VRadioGroup(
				VRadio().Value("1").Label("Tokyo"),
				VRadio().Value("2").Label("Canberra"),
				VRadio().Value("3").Label("Hangzhou"),
			).Label(field.Label).Value(field.Value(obj)).FieldName(field.Name)
		})
	ed.Field("FileInput1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VFileInput().Label(field.Label).Value(field.Value(obj)).FieldName(field.Name)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			fs := ctx.R.MultipartForm.File[field.Name]
			if len(fs) == 0 {
				return
			}
			f, err := fs[0].Open()
			if err != nil {
				panic(err)
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			obj.(*models.InputHarness).FileInput1 = fmt.Sprint(len(b))

			return
		})

	ed.Field("Combobox1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VCombobox().Label(field.Label).Value(field.Value(obj)).
				Attr(web.VFieldName(field.Name)...).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"})
		})

	ed.Field("Checkbox1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VCheckbox().Label(field.Label).
				Value(field.Value(obj)).InputValue(field.Value(obj)).
				//Attr("@change", web.Plaid().FieldValue(field.Name, web.Var("$event")).String()).
				Attr(web.VFieldName(field.Name)...)
		})

	ed.Field("Autocomplete1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VAutocomplete().Label(field.Label).
				Value(field.Value(obj)).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"}).
				//Attr("@change", web.Plaid().FieldValue(field.Name, web.Var("$event")).String()).
				FieldName(field.Name)
		})
}
