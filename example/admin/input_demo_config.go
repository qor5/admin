package admin

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/samber/lo"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configInputDemo(b *presets.Builder, db *gorm.DB) {
	inputDemo := b.Model(&models.InputDemo{})
	// MenuIcon("view_quilt")

	inputDemo.Listing().OrderableFields([]*presets.OrderableField{
		{
			FieldName: "ID",
			DBColumn:  "id",
		},
		{
			FieldName: "TextField1",
			DBColumn:  "text_field1",
		},
		{
			FieldName: "Switch1",
			DBColumn:  "switch1",
		},
	})

	ed := inputDemo.Editing(
		"TextField1",
		"TextArea1",
		"Switch1",
		"Slider1",
		"Select1",
		// "RangeSlider1",
		"Radio1",
		"FileInput1",
		"Combobox1",
		"Checkbox1",
		"Autocomplete1",
		"ButtonGroup1",
		"ChipGroup1",
		// "ItemGroup1",
		// "ListItemGroup1",
		// "SlideGroup1",
		// "ColorPicker1",
		"DatePicker1",
		"DatePickerMonth1",
		"TimePicker1",
		"MediaLibrary1",
		"SelectedCustomers",
	)

	// TextField1       string
	// TextArea1        string
	// Switch1          bool
	// Slider1          int
	// Select1          string
	// RangeSlider1     string
	// Radio1           string
	// FileInput1       string
	// Combobox1        string
	// Checkbox1        string
	// Autocomplete1    string
	// ButtonGroup1     string
	// ChipGroup1       string
	// ItemGroup1       string
	// ListItemGroup1   string
	// SlideGroup1      string
	// ColorPicker1     string
	// DatePicker1      string
	// DatePickerMonth1 string
	// TimePicker1      string
	// MediaLibrary1    media_library.MediaBox

	ed.Field("TextField1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VTextField().Label(field.Label).Attr(web.VField(field.Name, field.Value(obj))...)
		})

	ed.Field("TextArea1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VTextarea().Label(field.Label).Attr(web.VField(field.Name, field.Value(obj))...)
		})

	ed.Field("Switch1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSwitch().Label(field.Label).Value(field.Value(obj)).Attr(web.VField(field.Name, field.Value(obj))...)
		})

	ed.Field("Slider1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSlider().Label(field.Label).Attr(web.VField(field.Name, field.Value(obj))...)
		})

	ed.Field("Select1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSelect().
				Label(field.Label).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"}).Attr(web.VField(field.Name, field.Value(obj))...)
		})

	// ed.Field("RangeSlider1").
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
			).Label(field.Label).Attr(web.VField(field.Name, field.Value(obj))...)
		})
	ed.Field("FileInput1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VFileInput().Label(field.Label).Attr(web.VField(field.Name, field.Value(obj))...)
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
			defer f.Close()

			b, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			obj.(*models.InputDemo).FileInput1 = fmt.Sprint(len(b))

			return
		})

	ed.Field("Combobox1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VCombobox().Label(field.Label).
				Attr(web.VField(field.Name, field.Value(obj))...).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"})
		})

	ed.Field("Checkbox1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VCheckbox().Label(field.Label).
				ModelValue(field.Value(obj)).
				Attr(web.VField(field.Name, field.Value(obj))...)
		})

	ed.Field("Autocomplete1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VAutocomplete().Label(field.Label).
				Attr(web.VField(field.Name, field.Value(obj))...).
				Items([]string{"Tokyo", "Canberra", "Hangzhou"})
			// Attr("@change", web.Plaid().FieldValue(field.Name, web.Var("$event")).String()).
		})

	ed.Field("ButtonGroup1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VInput(

				VBtnToggle(
					VBtn("Left").Value("left").Class("deep-purple white--text"),
					VBtn("Center").Value("center").Class("deep-purple white--text"),
					VBtn("Right").Value("right").Class("deep-purple white--text"),
					VBtn("Justify").Value("justify").Class("deep-purple white--text"),
				).
					Class("pl-4").
					Attr(web.VField(field.Name, field.Value(obj))...),
			).Label(field.Label)
		})

	ed.Field("ChipGroup1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VInput(

				VChipGroup(
					VChip(h.Text("Left")).Filter(true).Value("left").ActiveClass("deep-purple white--text"),
					VChip(h.Text("Center")).Filter(true).Value("center").ActiveClass("deep-purple white--text"),
					VChip(h.Text("Right")).Filter(true).Value("right").ActiveClass("deep-purple white--text"),
					VChip(h.Text("Justify")).Filter(true).Value("justify").ActiveClass("deep-purple white--text"),
				).
					Class("pl-4").
					Attr(web.VField(field.Name, field.Value(obj))...),
			).Label(field.Label)
		})

	ed.Field("DatePicker1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return vx.Picker(VDatePicker()).FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("DatePickerMonth1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return vx.Picker(VDatePicker().Month(true)).FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("TimePicker1").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return vx.Picker(VTimePicker()).FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	ed.Field("MediaLibrary1").
		WithContextValue(
			media.MediaBoxConfig,
			&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*base.Size{
					"thumb": {
						Width:  400,
						Height: 300,
					},
					"main": {
						Width:  800,
						Height: 500,
					},
				},
			})

	configureDialogCustomerSelector(db, b)
	ed.Field("SelectedCustomers").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		compo, err := customerSelector(ctx, nil, []string{})
		if err != nil {
			panic(err)
		}
		return web.Portal(compo).Name(portalCustomerSelector)
	})
}

const (
	eventSelectCustomer           = "eventSelectCustomer"
	portalCustomerSelector        = "portalCustomerSelector"
	uriNameDialogCustomerSelector = "dialog-customer-selector"
)

func customerSelector(_ *web.EventContext, db *gorm.DB, selectedIds []string) (h.HTMLComponent, error) {
	var items []*models.Customer
	if len(selectedIds) > 0 {
		if err := db.Where("id IN (?)", selectedIds).Find(&items).Error; err != nil {
			return nil, errors.Wrap(err, "find customers")
		}
	}
	children := []h.HTMLComponent{
		VBtn("Select Customers").Class("mb-2").Attr("@click",
			web.Plaid().URL("/"+uriNameDialogCustomerSelector).
				EventFunc(actions.OpenListingDialog).
				Query("selected_ids", selectedIds).
				Go(),
		),
	}
	children = append(children, lo.Map(items, func(v *models.Customer, _ int) h.HTMLComponent {
		return h.Div().Children(
			h.Text(v.Name),
		)
	})...)

	return h.Div().Class("d-flex flex-column ga-2 align-start").Children(children...), nil
}

const paramSelectedIds = "selected_ids"

func configureDialogCustomerSelector(db *gorm.DB, pb *presets.Builder) {
	b := pb.Model(&models.Customer{}).URIName(uriNameDialogCustomerSelector).InMenu(false)
	registerEventSelectCustomer(db, pb)

	lb := b.Listing().DialogWidth("900px")
	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.RowMenu().Empty()
	lb.WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
		return in
	})
	lb.BulkAction("Confirm").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return VBtn("Confirm").Attr("@click", web.Plaid().
			EventFunc(eventSelectCustomer).
			Query(paramSelectedIds, web.Var("locals.selected_ids")).
			MergeQuery(true).
			Go(),
		)
	})
}

func registerEventSelectCustomer(db *gorm.DB, b *presets.Builder) {
	b.GetWebBuilder().RegisterEventFunc(eventSelectCustomer, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		selectedIds := strings.Split(strings.TrimSpace(ctx.R.FormValue(paramSelectedIds)), ",")
		selectedIds = lo.Filter(selectedIds, func(v string, _ int) bool { return v != "" })
		compo, err := customerSelector(ctx, db, selectedIds)
		if err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: portalCustomerSelector,
			Body: compo,
		})
		web.AppendRunScripts(&r, presets.CloseListingDialogVarScript)
		return
	})
}
